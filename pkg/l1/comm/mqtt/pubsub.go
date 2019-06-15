package mqtt

import (
	"container/list"
	"net/url"
	"strings"
	"sync"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/glog"
)

// Handler is the callback when a message is received.
type Handler func(topic string, payload []byte)

// Queue wraps MQTT client.
type Queue struct {
	Client       paho.Client
	TopicPrefix  string
	OnConnect    ConnectHandler
	OnDisconnect ConnectHandler

	subsLock     sync.RWMutex
	subs         map[string]*list.List
	wildcardSubs map[string]*list.List
}

// ConnectHandler is to handle connect/disconnect events.
type ConnectHandler func(*Queue)

// Subscription is a subscribed topic.
type Subscription struct {
	Token paho.Token

	queue    *Queue
	elm      *list.Element
	topic    string
	wildcard bool
	handler  Handler
}

// MatchTopic matches topic with pattern.
func MatchTopic(topic, pattern string) bool {
	tokensT, tokensP := strings.Split(topic, "/"), strings.Split(pattern, "/")
	if len(tokensP) > len(tokensT) {
		return false
	}
	for i, token := range tokensP {
		if token == "+" {
			continue
		}
		if token == "#" && i+1 == len(tokensP) {
			break
		}
		if token != tokensT[i] {
			return false
		}
	}
	return true
}

// ClientOptionsFromURL creates ClientOptions from URL.
func ClientOptionsFromURL(serverURL string) (*paho.ClientOptions, string, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, "", err
	}
	var server string
	if u.Scheme == "" || u.Scheme == "mqtt" {
		server = "tcp"
	} else {
		server = u.Scheme
	}
	server += "://" + u.Host

	topicPrefix := u.Path
	if strings.HasPrefix(topicPrefix, "/") {
		topicPrefix = topicPrefix[1:]
	}

	opts := paho.NewClientOptions()
	opts.AddBroker(server).
		SetAutoReconnect(true).
		SetCleanSession(true)
	if u.User != nil {
		opts.SetUsername(u.User.Username())
		if pwd, ok := u.User.Password(); ok {
			opts.SetPassword(pwd)
		}
	}

	if clientID := u.Query().Get("client-id"); clientID != "" {
		opts.SetClientID(clientID)
	}

	return opts, topicPrefix, nil
}

// NewQueue creates Queue.
func NewQueue(options *paho.ClientOptions, topicPrefix string) *Queue {
	q := &Queue{TopicPrefix: topicPrefix}
	options.SetOnConnectHandler(q.OnConnectHandler)
	options.SetConnectionLostHandler(q.ConnectionLostHandler)
	q.Client = paho.NewClient(options)
	return q
}

// NewQueueFromURL creates Queue from URL.
func NewQueueFromURL(brokerURL string) (*Queue, error) {
	opts, topicPrefix, err := ClientOptionsFromURL(brokerURL)
	if err != nil {
		return nil, err
	}
	return NewQueue(opts, topicPrefix), nil
}

// Connect connects the client.
func (q *Queue) Connect() paho.Token {
	return q.Client.Connect()
}

// Close implements io.Closer.
func (q *Queue) Close() error {
	q.Client.Disconnect(0)
	return nil
}

// Sub subscribes a topic
func (q *Queue) Sub(topic string, handler Handler) *Subscription {
	wildcard := strings.Contains(topic, "+") || strings.HasSuffix(topic, "#")
	var newSub bool
	q.subsLock.Lock()
	if q.subs == nil {
		q.subs = make(map[string]*list.List)
	}
	if q.wildcardSubs == nil {
		q.wildcardSubs = make(map[string]*list.List)
	}
	subs := q.subs
	if wildcard {
		subs = q.wildcardSubs
	}
	lst := subs[topic]
	if lst == nil {
		lst = list.New()
		subs[topic] = lst
		newSub = true
	}
	sub := &Subscription{
		queue:    q,
		topic:    topic,
		wildcard: wildcard,
		handler:  handler,
	}
	sub.elm = lst.PushBack(sub)
	q.subsLock.Unlock()

	if newSub {
		if glog.V(2) {
			glog.Infof("SUB %q", q.TopicPrefix+topic)
		}
		sub.Token = q.Client.Subscribe(q.TopicPrefix+topic, 0, q.dispatch)
	}
	return sub
}

// Pub publishes to a topic.
func (q *Queue) Pub(topic string, payload []byte) paho.Token {
	return q.PubWith(topic, payload, 0, false)
}

// PubWith publishes with QoS and retain settings.
func (q *Queue) PubWith(topic string, payload []byte, qos byte, retain bool) paho.Token {
	return q.Client.Publish(q.TopicPrefix+topic, qos, retain, payload)
}

// Resubscribe is used in OnConnect handler to subscribe all existing topics.
func (q *Queue) Resubscribe() paho.Token {
	filters := make(map[string]byte)
	q.subsLock.RLock()
	for topic := range q.subs {
		filters[q.TopicPrefix+topic] = 0
	}
	for topic := range q.wildcardSubs {
		filters[q.TopicPrefix+topic] = 0
	}
	q.subsLock.RUnlock()
	if len(filters) > 0 {
		if glog.V(2) {
			for key := range filters {
				glog.Infof("SUB %q", key)
			}
		}
		return q.Client.SubscribeMultiple(filters, q.dispatch)
	}
	return &paho.DummyToken{}
}

// OnConnectHandler is the default implementation of paho.OnConnectHandler.
func (q *Queue) OnConnectHandler(paho.Client) {
	glog.Info("connected")
	q.Resubscribe()
	if h := q.OnConnect; h != nil {
		h(q)
	}
}

// ConnectionLostHandler is the default implementation of paho.ConnectLostHandler.
func (q *Queue) ConnectionLostHandler(c paho.Client, err error) {
	glog.Warningf("connection lost: %v", err)
	if h := q.OnDisconnect; h != nil {
		h(q)
	}
}

func (q *Queue) dispatch(c paho.Client, msg paho.Message) {
	if topic := msg.Topic(); strings.HasPrefix(topic, q.TopicPrefix) {
		glog.V(2).Infof("RCV %q", topic)
		topic = topic[len(q.TopicPrefix):]
		var handlers []Handler
		q.subsLock.RLock()
		if lst := q.subs[topic]; lst != nil {
			handlers = make([]Handler, 0, lst.Len())
			for elm := lst.Front(); elm != nil; elm = elm.Next() {
				handlers = append(handlers, elm.Value.(*Subscription).handler)
			}
		}
		for key, lst := range q.wildcardSubs {
			if MatchTopic(topic, key) {
				for elm := lst.Front(); elm != nil; elm = elm.Next() {
					handlers = append(handlers, elm.Value.(*Subscription).handler)
				}
			}
		}
		q.subsLock.RUnlock()
		payload := msg.Payload()
		for _, h := range handlers {
			h(topic, payload)
		}
	}
}

// Close unsubscribes a handler.
func (s *Subscription) Close() error {
	var unsub bool
	s.queue.subsLock.Lock()
	lst := s.queue.subs[s.topic]
	if lst != nil {
		lst.Remove(s.elm)
		if unsub = lst.Len() == 0; unsub {
			delete(s.queue.subs, s.topic)
		}
	} else if lst = s.queue.wildcardSubs[s.topic]; lst != nil {
		lst.Remove(s.elm)
		if unsub = lst.Len() == 0; unsub {
			delete(s.queue.wildcardSubs, s.topic)
		}
	}
	s.queue.subsLock.Unlock()
	if unsub {
		glog.V(2).Infof("UNSUB %q", s.topic)
		token := s.queue.Client.Unsubscribe(s.queue.TopicPrefix + s.topic)
		token.Wait()
		return token.Error()
	}
	return nil
}

package mqtt

import (
	"context"
	"strings"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"github.com/robotalks/robo.go/pkg/l1"
	"github.com/robotalks/robo.go/pkg/l1/comm"
)

// Connector implements l1.Connector using MQTT.
type Connector struct {
	DiscoverTimeout time.Duration

	options     *paho.ClientOptions
	topicPrefix string
}

// DefaultDiscoverTimeout defines the default timeout value of discovery.
const DefaultDiscoverTimeout = 500 * time.Millisecond

// NewConnector creates a Connector.
func NewConnector(brokerURL string) (*Connector, error) {
	opts, topicPrefix, err := ClientOptionsFromURL(brokerURL)
	if err != nil {
		return nil, err
	}
	return &Connector{
		DiscoverTimeout: DefaultDiscoverTimeout,
		options:         opts,
		topicPrefix:     topicPrefix,
	}, nil
}

// Discover implements Connector.
func (c *Connector) Discover(ctx context.Context) (res []l1.ControllerInfo, err error) {
	q := NewQueue(c.options, c.topicPrefix)
	q.Connect()
	defer q.Close()
	resCh := make(chan l1.ControllerInfo, 1)
	q.Sub("+/+/meta", Handler(func(topic string, payload []byte) {
		items := strings.Split(topic, "/")
		if len(items) == 3 {
			select {
			case resCh <- l1.ControllerInfo{Ref: l1.ControllerRef{Type: items[0], ID: items[1]}}:
			case <-time.After(time.Second):
			}
		}
	}))

	dur := c.DiscoverTimeout
	if dur == 0 {
		dur = DefaultDiscoverTimeout
	}
	timeout := time.After(dur)
	for {
		select {
		case info := <-resCh:
			res = append(res, info)
		case <-timeout:
			return
		case <-ctx.Done():
			err = ctx.Err()
			return
		}
	}
}

// Connect implements Connector.
func (c *Connector) Connect(ctx context.Context, ref l1.ControllerRef) (l1.ControllerConn, error) {
	conn := &ControllerConn{
		Queue: NewQueue(c.options, c.topicPrefix),
	}
	conn.Init(NewPacketReadWriter(conn.Queue).ForConnector(ref))
	token := conn.Queue.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		return nil, err
	}
	return conn, nil
}

// ControllerConn implements ControllerConn using MQTT.
type ControllerConn struct {
	comm.ControllerConn
	Queue *Queue
}

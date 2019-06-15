package comm

import (
	"container/list"
	"context"
	"sync"
	"time"

	fx "github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/l1"
	"github.com/robotalks/robo.go/pkg/l1/msgs"
)

// ControllerConn provides base implementation for l1.ControllerConn using Pipe.
type ControllerConn struct {
	Expiration time.Duration

	pipe     Pipe
	seq      uint32
	commands list.List
	seqMap   map[uint32]*commandFuture
	lock     sync.Mutex
}

// DefaultCommandExpiration is the default expiration expecting a result.
const DefaultCommandExpiration = 1 * time.Second

// Init initializes ControllerConn with defaults.
func (c *ControllerConn) Init(rw PacketReadWriter) {
	c.Expiration = DefaultCommandExpiration
	c.pipe.ReadWriter = rw
	c.pipe.Handler = msgs.HandleTypedMsgFunc(c.handleTypedMsg)
	c.seqMap = make(map[uint32]*commandFuture)
}

// DoCommand implements ControllerConn.
func (c *ControllerConn) DoCommand(msg fx.Message) l1.CommandFuture {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.seq++
	if c.seq == 0 {
		c.seq++
	}
	f := &commandFuture{
		seq:      c.seq,
		expireAt: time.Now().Add(c.Expiration),
		result:   make(chan l1.Result, 1),
	}
	if err := c.pipe.SendCommandMsg(msg, f.seq); err != nil {
		f.result <- l1.Result{Err: err}
		return f
	}
	f.elem = c.commands.PushBack(f)
	c.seqMap[f.seq] = f
	return f
}

// AddToLoop implements LoopAdder.
func (c *ControllerConn) AddToLoop(l *fx.Loop) {
	l.Add(&c.pipe)
	l.AddController(fx.PrLvIdle, fx.ControlFunc(c.purgeExpired))
}

func (c *ControllerConn) handleTypedMsg(ctx context.Context, msg fx.Message, typed *msgs.Typed) error {
	if typed.IsEvent() {
		loopCtl := fx.LoopCtlFrom(ctx)
		loopCtl.PostMessage(msg)
		loopCtl.TriggerNext()
		return nil
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	f := c.seqMap[typed.Sequence]
	if f == nil {
		return nil
	}
	c.commands.Remove(f.elem)
	delete(c.seqMap, typed.Sequence)
	result := l1.Result{Msg: msg}
	if cmdErr, ok := msg.(*msgs.CommandErr); ok {
		result.Err = cmdErr
	}
	f.result <- result
	close(f.result)
	return nil
}

func (c *ControllerConn) purgeExpired(cc fx.ControlContext) error {
	now := time.Now()
	c.lock.Lock()
	defer c.lock.Unlock()
	for c.commands.Len() > 0 {
		elem := c.commands.Front()
		f := elem.Value.(*commandFuture)
		if f.expireAt.After(now) {
			break
		}
		c.commands.Remove(elem)
		delete(c.seqMap, f.seq)
		f.result <- l1.Result{Err: context.DeadlineExceeded}
		close(f.result)
	}
	return nil
}

type commandFuture struct {
	seq      uint32
	expireAt time.Time
	elem     *list.Element
	result   chan l1.Result
}

func (c *commandFuture) ResultChan() <-chan l1.Result {
	return c.result
}

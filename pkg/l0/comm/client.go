package comm

import (
	"context"
	"sync"
)

// Result is the result of a command using Do.
type Result struct {
	Err  error
	Code byte
	Data []byte
}

// Client provides client side operations over FIFO.
type Client struct {
	fifo     *FIFO
	eventCh  chan *Packet
	stateCh  chan SyncState
	cmdsHead *Command
	cmdsTail *Command
	cmdsLock sync.Mutex
}

// Command represents a pending command waiting for reply.
type Command struct {
	requestSeq PacketSeq
	resultCh   chan Result
	next       *Command
}

// RequestSeq returns the request packet seq.
func (c *Command) RequestSeq() PacketSeq {
	return c.requestSeq
}

// ResultChan returns the chan to retrieve result.
func (c *Command) ResultChan() <-chan Result {
	return c.resultCh
}

// NewClient creates client and wraps the fifo.
func NewClient(fifo *FIFO) *Client {
	c := &Client{
		fifo:    fifo,
		eventCh: make(chan *Packet, 1),
		stateCh: make(chan SyncState, 1),
	}
	c.fifo.Handler = c
	c.fifo.Notifier = StateChangedFunc(func(ctx context.Context, state SyncState) {
		c.stateCh <- state
	})
	return c
}

// FIFO gets wrapped FIFO.
func (c *Client) FIFO() *FIFO {
	return c.fifo
}

// StateChan retrieves the state reporting chan.
func (c *Client) StateChan() <-chan SyncState {
	return c.stateCh
}

// EventChan retrieves the event reporting chan.
func (c *Client) EventChan() <-chan *Packet {
	return c.eventCh
}

// DoWith sends a command and expects a result in the provided chan.
func (c *Client) DoWith(pkt *Packet, ch chan Result) *Command {
	cmd := &Command{resultCh: ch}

	c.cmdsLock.Lock()
	defer c.cmdsLock.Unlock()
	err := c.fifo.Send(pkt)
	cmd.requestSeq = pkt.Seq
	if err != nil {
		cmd.resultCh <- Result{Err: err}
		return cmd
	}
	if c.cmdsHead == nil {
		c.cmdsHead = cmd
	} else {
		c.cmdsTail.next = cmd
	}
	c.cmdsTail = cmd
	return cmd
}

// Do sends a command and returns a Command for result.
func (c *Client) Do(pkt *Packet) *Command {
	return c.DoWith(pkt, make(chan Result, 1))
}

// HandlePacket implements PacketHandler.
func (c *Client) HandlePacket(ctx context.Context, pkt *Packet) {
	if pkt.Code&0x80 != 0 {
		c.eventCh <- pkt
		return
	}
	if len(pkt.Data) == 0 {
		// invalid response packet.
		return
	}
	seq := PacketSeq(pkt.Data[0])
	if !seq.IsValid() {
		// invalid sequence.
		return
	}
	c.cmdsLock.Lock()
	head := c.cmdsHead
	curr := c.cmdsHead
	for ; curr != nil; curr = curr.next {
		if curr.requestSeq == seq {
			if c.cmdsHead = curr.next; c.cmdsHead == nil {
				c.cmdsTail = nil
			}
			curr.next = nil
			break
		}
	}
	c.cmdsLock.Unlock()
	if curr == nil {
		return
	}
	for ; head != curr; head = head.next {
		head.resultCh <- Result{Err: ErrNoReply}
	}
	if pkt.Code&1 != 0 {
		curr.resultCh <- Result{Err: &CommandError{Code: pkt.Code & 0x7e}}
	} else {
		curr.resultCh <- Result{Code: pkt.Code & 0x7e, Data: pkt.Data[1:]}
	}
}

// Run wraps FIFO.Run to implement Runnable.
func (c *Client) Run(ctx context.Context) error {
	return c.fifo.Run(ctx)
}

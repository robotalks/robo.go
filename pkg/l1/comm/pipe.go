package comm

import (
	"context"
	"io"
	"sync"

	fx "github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/l1/msgs"
)

// Pipe is a bi-directional pipe for messages.
type Pipe struct {
	ReadWriter PacketReadWriter
	Handler    msgs.TypedMsgHandler

	sendLock sync.Mutex
}

// NewPipe creates a Pipe with given PacketReadWriter.
func NewPipe(rw PacketReadWriter) *Pipe {
	return &Pipe{ReadWriter: rw}
}

// SendCommandMsg sends a message which must be a command.
func (p *Pipe) SendCommandMsg(msg fx.Message, seq uint32) error {
	typed, err := msgs.TypedFrom(msg)
	if err != nil {
		panic(err)
	}
	if !typed.IsCommand() {
		panic("message is not a command")
	}
	typed.Sequence = seq
	return p.SendTyped(typed)
}

// SendEventMsg sends a message which must be an event.
func (p *Pipe) SendEventMsg(msg fx.Message) error {
	typed, err := msgs.TypedFrom(msg)
	if err != nil {
		panic(err)
	}
	if !typed.IsEvent() {
		panic("message is not an event")
	}
	return p.SendTyped(typed)
}

// SendTyped send a Typed message.
func (p *Pipe) SendTyped(typed *msgs.Typed) error {
	pkt, err := typed.Encode()
	if err != nil {
		return err
	}
	p.sendLock.Lock()
	defer p.sendLock.Unlock()
	return p.ReadWriter.WritePacket(pkt)
}

// Run implements Runnable.
func (p *Pipe) Run(ctx context.Context) error {
	defer p.Close()
	for {
		pkt, err := p.ReadWriter.ReadPacket()
		if err != nil {
			return err
		}
		typed, err := msgs.DecodeTyped(pkt)
		if err != nil {
			return err
		}
		msg, err := typed.Decode()
		if err != nil {
			// If it's command, simply replies a CommandErr.
			if typed.IsCommand() {
				if err = p.SendCommandMsg(msgs.NewCommandErr(err), typed.Sequence); err != nil {
					return err
				}
			}
			// otherwise, simply ignored.
			continue
		}
		if h := p.Handler; h != nil {
			err = h.HandleTypedMsg(ctx, msg, typed)
		}
		if err != nil {
			return err
		}
	}
}

// Close implements Closer.
func (p *Pipe) Close() error {
	if closer, ok := p.ReadWriter.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// AddToLoop implements LoopAdder.
func (p *Pipe) AddToLoop(loop *fx.Loop) {
	if adder, ok := p.ReadWriter.(fx.LoopAdder); ok {
		loop.Add(adder)
	} else if runnable, ok := p.ReadWriter.(fx.Runnable); ok {
		loop.AddRunnable(runnable)
	}
	loop.AddRunnable(p)
}

package comm

import (
	"context"

	fx "github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/l1"
	"github.com/robotalks/robo.go/pkg/l1/msgs"
)

// Registrar implements Registrar with Pipe and integrated with Loop.
type Registrar struct {
	pipe Pipe
}

// Init initializes the Registrar with defaults.
func (r *Registrar) Init(rw PacketReadWriter) {
	r.pipe.ReadWriter = rw
	r.pipe.Handler = msgs.HandleTypedMsgFunc(func(ctx context.Context, msg fx.Message, typed *msgs.Typed) error {
		loopCtl := fx.LoopCtlFrom(ctx)
		switch typed.Kind() {
		case msgs.TypeIDKindCommand:
			loopCtl.PostMessage(&l1.CommandMsg{Command: &command{seq: typed.Sequence, msg: msg, pipe: &r.pipe}})
			loopCtl.TriggerNext()
		case msgs.TypeIDKindEvent:
			loopCtl.PostMessage(msg)
			loopCtl.TriggerNext()
		}
		return nil
	})
}

// SendEvent implements Registrar.
func (r *Registrar) SendEvent(ctx context.Context, msg fx.Message) error {
	return r.pipe.SendEventMsg(msg)
}

// AddToLoop implements LoopAdder.
func (r *Registrar) AddToLoop(loop *fx.Loop) {
	loop.Add(&r.pipe)
}

type command struct {
	seq  uint32
	msg  fx.Message
	pipe *Pipe
}

func (c *command) Msg() fx.Message {
	return c.msg
}

func (c *command) Done(msg fx.Message) error {
	return c.pipe.SendCommandMsg(msg, c.seq)
}

// RegistrarMux registers L1 controller with multiple Registrars.
type RegistrarMux struct {
	Registrars []l1.Registrar
}

// SendEvent implements Registrar.
func (r *RegistrarMux) SendEvent(ctx context.Context, msg fx.Message) error {
	var errs fx.AggregatedError
	for _, reg := range r.Registrars {
		errs.Add(reg.SendEvent(ctx, msg))
	}
	return errs.Aggregate()
}

// AddToLoop implements LoopAdder.
func (r *RegistrarMux) AddToLoop(l *fx.Loop) {
	for _, reg := range r.Registrars {
		if adder, ok := reg.(fx.LoopAdder); ok {
			l.Add(adder)
		}
	}
}

// Add adds more registrars.
func (r *RegistrarMux) Add(regs ...l1.Registrar) {
	r.Registrars = append(r.Registrars, regs...)
}

// UnsupportedCommands replies left-over commands as unsupported.
type UnsupportedCommands struct {
}

// Control implements Controller.
func (c *UnsupportedCommands) Control(cc fx.ControlContext) error {
	cc.Messages().ProcessMessages(fx.ProcessMessageFunc(func(mctx fx.MessageProcessingContext) {
		if cmdMsg, ok := mctx.CurrentMessage().(*l1.CommandMsg); ok {
			mctx.MessageTaken()
			cmdMsg.Command.Done(msgs.NewCommandErr(msgs.ErrUnsupportedCommand))
		}
	}))
	return nil
}

// AddToLoop implements LoopAdder.
func (c *UnsupportedCommands) AddToLoop(loop *fx.Loop) {
	loop.AddController(fx.PrLvIdle, c)
}

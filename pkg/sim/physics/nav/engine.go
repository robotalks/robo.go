package nav

import (
	"time"

	fx "github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/l1"
	"github.com/robotalks/robo.go/pkg/l1/msgs"
	"github.com/robotalks/robo.go/pkg/sim"
	"github.com/robotalks/robo.go/pkg/sim/physics"
)

// Engine implements physics.Nav2D
type Engine struct {
	Object sim.Placeable2D
	Caps   msgs.Nav2DCaps

	state state
}

type state interface {
	estimate(now time.Time) (sim.Pose2D, state)
}

// New creates the engine.
func New(obj sim.Placeable2D) *Engine {
	return &Engine{Object: obj}
}

// CapsQuery executes Nav2DCapsQuery command.
func (e *Engine) CapsQuery(ctx physics.Context, msg *msgs.Nav2DCapsQuery) *msgs.Nav2DCaps {
	return &e.Caps
}

// Drive executes Nav2DDrive command.
func (e *Engine) Drive(ctx physics.Context, msg *msgs.Nav2DDrive) {
	e.state = newDriveState(e.state, e.estimatePose(ctx), ctx.Time(), msg)
}

// Turn executes Nav2DTurn command.
func (e *Engine) Turn(ctx physics.Context, msg *msgs.Nav2DTurn) {
	e.state = newTurnState(e.estimatePose(ctx), ctx.Time(), msg)
}

// AddToLoop implements LoopAdder.
func (e *Engine) AddToLoop(l *fx.Loop) {
	l.AddController(fx.PrLvControl, fx.ControlFunc(e.HandleCommand))
	l.AddController(fx.PrLvAcuate, fx.ControlFunc(e.Execute))
}

// HandleCommand is a controller processing commands.
func (e *Engine) HandleCommand(cc fx.ControlContext) error {
	cc.Messages().ProcessMessages(fx.ProcessMessageFunc(func(mctx fx.MessageProcessingContext) {
		if cmdMsg, ok := mctx.CurrentMessage().(*l1.CommandMsg); ok {
			switch m := cmdMsg.Command.Msg().(type) {
			case *msgs.Nav2DCapsQuery:
				mctx.MessageTaken()
				cmdMsg.Command.Done(e.CapsQuery(cc, m))
			case *msgs.Nav2DDrive:
				mctx.MessageTaken()
				e.Drive(cc, m)
				cmdMsg.Command.Done(msgs.NewCommandOK())
			case *msgs.Nav2DTurn:
				mctx.MessageTaken()
				e.Turn(cc, m)
				cmdMsg.Command.Done(msgs.NewCommandOK())
			}
		}
	}))
	return nil
}

// Execute is a controller for acuation.
func (e *Engine) Execute(ctx fx.ControlContext) error {
	if s := e.state; s != nil {
		var pose sim.Pose2D
		pose, e.state = s.estimate(ctx.Time())
		e.Object.SetPose2D(pose)
	}
	return nil
}

func (e *Engine) estimatePose(ctx physics.Context) (pose sim.Pose2D) {
	if s := e.state; s != nil {
		pose, e.state = s.estimate(ctx.Time())
		pose = e.Object.SetPose2D(pose)
	} else {
		pose = e.Object.Position2D()
	}
	return
}

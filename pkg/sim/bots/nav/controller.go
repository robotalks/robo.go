package nav

import (
	fx "github.com/robotalks/robo.go/pkg/framework"
	env "github.com/robotalks/robo.go/pkg/l1/env/controller"
	"github.com/robotalks/robo.go/pkg/sim"
	"github.com/robotalks/robo.go/pkg/sim/physics/nav"
)

// Controller is the L1 controller.
type Controller struct {
	Env *env.Env

	Outline sim.Rect
	Pose    sim.Pose2D
	Nav     *nav.Engine

	sim.ObjectsChangeCaster

	ref     string
	changes int
}

// NewController creates the controller.
func NewController(e *env.Env) *Controller {
	c := &Controller{
		Env:     e,
		changes: 1, // send initial object change.
	}
	c.Nav = nav.New(c)
	return c
}

// Name implements Named.
func (c *Controller) Name() string {
	return c.Env.Config.Info.Ref.Name()
}

// AddToLoop implements LoopAdder.
func (c *Controller) AddToLoop(l *fx.Loop) {
	l.Add(c.Nav)
	l.AddController(fx.PrLvPostProc, fx.ControlFunc(c.NotifyChanges))
}

// OutlineRect implements Rectangular.
func (c *Controller) OutlineRect() sim.Rect {
	return c.Outline
}

// Position2D implements Placeable2D.
func (c *Controller) Position2D() sim.Pose2D {
	return c.Pose
}

// SetPose2D implements Placeable2D.
func (c *Controller) SetPose2D(pose sim.Pose2D) sim.Pose2D {
	c.Pose = pose
	c.changes = 1
	return c.Pose
}

// NotifyChanges notifies object changes.
func (c *Controller) NotifyChanges(cc fx.ControlContext) error {
	changes := c.changes
	c.changes = 0
	if changes > 0 {
		c.ObjectsChanged(cc, c)
	}
	return nil
}

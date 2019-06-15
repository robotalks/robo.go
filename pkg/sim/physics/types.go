package physics

import (
	"context"

	fx "github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/l1/msgs"
)

// Context provides the simulation context.
type Context interface {
	fx.TimeSource
	Context() context.Context
}

// Nav2D simulates navigation.
type Nav2D interface {
	Drive(Context, *msgs.Nav2DDrive)
	Turn(Context, *msgs.Nav2DTurn)
}

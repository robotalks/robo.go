package physics

import (
	"context"
	"time"
)

// NowContext wraps context.Context with current time.
type NowContext struct {
	now time.Time
	ctx context.Context
}

// Now creates NowContext.
func Now(ctx context.Context) Context {
	return &NowContext{now: time.Now(), ctx: ctx}
}

// Time implements TimeSource.
func (c NowContext) Time() time.Time {
	return c.now
}

// Context implements Context.
func (c NowContext) Context() context.Context {
	return c.ctx
}

package sim

import (
	fx "github.com/robotalks/robo.go/pkg/framework"
)

// Size2D defines the rectangular size in 2D.
type Size2D struct {
	CX, CY float64
}

// Size defines the cube size in 3D.
type Size struct {
	CX, CY, CZ float64
}

// Pos2D defines the position in 2D.
type Pos2D struct {
	X, Y float64
}

// Pos defines the position in 3D.
type Pos struct {
	X, Y, Z float64
}

// Rect defines a rectangle in 2D.
type Rect struct {
	Pos2D
	Size2D
}

// Cube defines a cube in 3D.
type Cube struct {
	Pos
	Size
}

// Pose2D defines the pose in 2D.
type Pose2D struct {
	Pos2D
	Orientation Angle
}

// Angle is the common representation of angle,
// supporting multiple units.
type Angle float64

// Rectangular object provides an rectangluar outline dimension.
type Rectangular interface {
	OutlineRect() Rect
}

// Positionable2D object maintains a 2D position.
type Positionable2D interface {
	Position2D() Pose2D
}

// Placeable2D object can be moved with a new pose on a 2D plane.
type Placeable2D interface {
	Positionable2D
	SetPose2D(Pose2D) Pose2D
}

// Object represents an object in the world.
type Object interface {
	fx.Named
}

// ObjectsChangeListener listens for object changes.
type ObjectsChangeListener interface {
	ObjectsChanged(fx.ControlContext, ...Object)
	ObjectsRemoved(fx.ControlContext, ...Object)
}

// ObjectsChangeSubscriber subscribes objects change notifications.
type ObjectsChangeSubscriber interface {
	SubscribeObjectsChange(ObjectsChangeListener)
}

// Add is a helper to add Pos2D.
func (p Pos2D) Add(p1 Pos2D) Pos2D {
	return Pos2D{X: p.X + p1.X, Y: p.Y + p1.Y}
}

// OffsetBy performs Add in-place.
func (p *Pos2D) OffsetBy(p1 Pos2D) *Pos2D {
	p.X += p1.X
	p.Y += p1.Y
	return p
}

// Add is a helper to add Pos.
func (p Pos) Add(p1 Pos) Pos {
	return Pos{X: p.X + p1.X, Y: p.Y + p1.Y, Z: p.Z + p1.Z}
}

// OffsetBy performs Add in-place.
func (p *Pos) OffsetBy(p1 Pos) *Pos {
	p.X += p1.X
	p.Y += p1.Y
	p.Z += p1.Z
	return p
}

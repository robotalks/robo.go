package see

import (
	"strings"

	"github.com/robotalks/robo.go/pkg/sim"
)

// VisibleObject is an object which can be visualized.
type VisibleObject interface {
	sim.Object
	sim.Rectangular
	sim.Positionable2D
}

// Object is the data model used to represents an object.
type Object map[string]interface{}

// Rect is object rect area.
type Rect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

// Pos is a position.
type Pos struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ObjectMapper maps VisibleObject into Object data model.
type ObjectMapper interface {
	MapObject(VisibleObject) []Object
}

// MapObjectFunc is the func form of ObjectMapper.
type MapObjectFunc func(VisibleObject) []Object

// MapObject implements ObjectMapper.
func (f MapObjectFunc) MapObject(obj VisibleObject) []Object {
	return f(obj)
}

// Message is the message for see.
type Message struct {
	Action   string `json:"action"`
	Object   Object `json:"object,omitempty"`
	RemoveID string `json:"id,omitempty"`
}

// Actions
const (
	ActionReset  = "reset"
	ActionObject = "object"
	ActionRemove = "remove"
)

// Properties
const (
	PropID     = "id"
	PropType   = "type"
	PropRect   = "rect"
	PropOrigin = "origin"
	PropRadius = "radius"
	PropRotate = "rotate"
	PropStyle  = "style"
	PropStyles = "styles"
)

// ObjectID converts object name to ID.
func ObjectID(name string) string {
	return strings.Replace(name, "/", ".", -1)
}

// NewObject creates Object.
func NewObject(typ, id string) Object {
	o := make(Object)
	o[PropID] = id
	o[PropType] = typ
	return o
}

// ObjectFrom constructs an object from VisibleObject.
func ObjectFrom(typ string, vo VisibleObject) Object {
	rc, po := vo.OutlineRect(), vo.Position2D()
	rad := rc.CX
	if rc.CY > rad {
		rad = rc.CY
	}
	return NewObject(typ, ObjectID(vo.Name())).
		At(po.X, po.Y).
		Radius(rad).
		Rotate(po.Orientation.Degrees())
}

// Rc sets rect.
func (o Object) Rc(x, y, w, h float64) Object {
	o[PropRect] = &Rect{X: x, Y: y, W: w, H: h}
	return o
}

// At sets origin.
func (o Object) At(x, y float64) Object {
	o[PropOrigin] = &Pos{X: x, Y: y}
	return o
}

// Radius sets radius.
func (o Object) Radius(r float64) Object {
	o[PropRadius] = r
	return o
}

// Rotate sets rotate.
func (o Object) Rotate(deg float64) Object {
	o[PropRotate] = deg
	return o
}

// With sets a custom property.
func (o Object) With(key string, val interface{}) Object {
	o[key] = val
	return o
}

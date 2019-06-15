package device

import "io"

// Event defines the base event interface.
type Event interface {
	// IsInit indicates this is the init state.
	IsInit() bool
	// Index returns either Axis or Button index.
	Index() int
}

// AxisEvent represents the change on an axis.
type AxisEvent interface {
	Event
	Value() int
}

// ButtonEvent represents the change on a button.
type ButtonEvent interface {
	Event
	Pressed() bool
}

// Device represents an opened joystick.
type Device interface {
	io.Closer
	// Index returns the index of the device on the system.
	Index() int
	// Name returns the name of the device.
	Name() string
	// AxisCount returns the number of Axis on the device.
	AxisCount() int
	// ButtonCount returns the number of buttons on the device.
	ButtonCount() int
	// ReadEvent reads one event from the device.
	ReadEvent() (Event, error)
}

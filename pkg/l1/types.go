package l1

import (
	"context"

	fx "github.com/robotalks/robo.go/pkg/framework"
)

// Registrar registers a robot (L1 controller) to an registry.
// It integrates with framework and helps an L1 controller to
// easily process messages.
type Registrar interface {
	// SendEvent sends an event to L2.
	SendEvent(context.Context, fx.Message) error
}

// Command represents a received command to be processed.
type Command interface {
	Msg() fx.Message
	Done(fx.Message) error
}

// CommandMsg wraps a Command as a Message.
type CommandMsg struct {
	Command Command
}

// NewMessage implements Message.
func (m *CommandMsg) NewMessage() fx.Message { return &CommandMsg{} }

// ControllerRef is a reference to an L1 controller.
type ControllerRef struct {
	// Type is controller type (robot type).
	Type string
	// ID is unique ID of the device.
	ID string
}

// Name retrieves the name from ref.
func (r ControllerRef) Name() string {
	return r.Type + "/" + r.ID
}

// IsValid indicates ControllerRef is valid.
func (r ControllerRef) IsValid() bool {
	return r.Type != "" && r.ID != ""
}

// ControllerMeta provides metadata for L1 controller.
type ControllerMeta struct {
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// ControllerInfo provides information of an L1 controller.
type ControllerInfo struct {
	Ref  ControllerRef
	Meta ControllerMeta
}

// Connector is used by L2 components to connect to an L1 controller.
type Connector interface {
	// Discover enumerates registered controllers.
	Discover(context.Context) ([]ControllerInfo, error)
	// Connect connects to the specified controller.
	Connect(context.Context, ControllerRef) (ControllerConn, error)
}

// ControllerConn is the connection to a controller.
type ControllerConn interface {
	// DoCommand executes a command.
	DoCommand(fx.Message) CommandFuture
}

// Result represents result of a command.
type Result struct {
	Msg fx.Message
	Err error
}

// CommandFuture is the future of sent command.
type CommandFuture interface {
	ResultChan() <-chan Result
}

package msgs

import (
	"errors"

	"github.com/golang/protobuf/proto"

	fx "github.com/robotalks/robo.go/pkg/framework"
	pb "github.com/robotalks/robo.go/pkg/proto/robo/l1/v1"
)

// CommandOK is the generic reply indicating success for commands.
type CommandOK struct {
	pb.CommandOK
}

// NewCommandOK creates a CommandOK.
func NewCommandOK() *CommandOK {
	return &CommandOK{}
}

// NewMessage implements Message.
func (m *CommandOK) NewMessage() fx.Message { return &CommandOK{} }

// TypeID implements SerializableMessage.
func (m *CommandOK) TypeID() uint32 { return CommandOKTypeID }

// Serializable implements SerializableMessage.
func (m *CommandOK) Serializable() proto.Message { return &m.CommandOK }

// CommandErr is the generic message representing command error.
type CommandErr struct {
	pb.CommandErr
}

// NewCommandErr creates a CommandErr from an error.
func NewCommandErr(err error) *CommandErr {
	return NewCommandErrFromMsg(err.Error())
}

// NewCommandErrFromMsg creates a CommandErr.
func NewCommandErrFromMsg(message string) *CommandErr {
	return &CommandErr{
		CommandErr: pb.CommandErr{
			Message: message,
		},
	}
}

// NewMessage implements Message.
func (m *CommandErr) NewMessage() fx.Message { return &CommandErr{} }

// TypeID implements SerializableMessage.
func (m *CommandErr) TypeID() uint32 { return CommandErrTypeID }

// Serializable implements SerializableMessage.
func (m *CommandErr) Serializable() proto.Message { return &m.CommandErr }

// Error implements error.
func (m *CommandErr) Error() string { return m.Message }

// Nav2DCapsQuery command.
type Nav2DCapsQuery struct {
	pb.Nav2DCapsQuery
}

// NewMessage implements Message.
func (m *Nav2DCapsQuery) NewMessage() fx.Message { return &Nav2DCapsQuery{} }

// TypeID implements SerializableMessage.
func (m *Nav2DCapsQuery) TypeID() uint32 { return Nav2DCapsQueryTypeID }

// Serializable implements SerializableMessage.
func (m *Nav2DCapsQuery) Serializable() proto.Message { return &m.Nav2DCapsQuery }

// Nav2DCaps response.
type Nav2DCaps struct {
	pb.Nav2DCaps
}

// NewMessage implements Message.
func (m *Nav2DCaps) NewMessage() fx.Message { return &Nav2DCaps{} }

// TypeID implements SerializableMessage.
func (m *Nav2DCaps) TypeID() uint32 { return Nav2DCapsTypeID }

// Serializable implements SerializableMessage.
func (m *Nav2DCaps) Serializable() proto.Message { return &m.Nav2DCaps }

// Nav2DDrive command.
type Nav2DDrive struct {
	pb.Nav2DDrive
}

// NewMessage implements Message.
func (m *Nav2DDrive) NewMessage() fx.Message { return &Nav2DDrive{} }

// TypeID implements SerializableMessage.
func (m *Nav2DDrive) TypeID() uint32 { return Nav2DDriveTypeID }

// Serializable implements SerializableMessage.
func (m *Nav2DDrive) Serializable() proto.Message { return &m.Nav2DDrive }

// Nav2DTurn command.
type Nav2DTurn struct {
	pb.Nav2DTurn
}

// NewMessage implements Message.
func (m *Nav2DTurn) NewMessage() fx.Message { return &Nav2DTurn{} }

// TypeID implements SerializableMessage.
func (m *Nav2DTurn) TypeID() uint32 { return Nav2DTurnTypeID }

// Serializable implements SerializableMessage.
func (m *Nav2DTurn) Serializable() proto.Message { return &m.Nav2DTurn }

// TypeID Groups
const (
	GroupCommand uint32 = 0x00000000
	GroupNav2D   uint32 = 0x00020000
	GroupCustom  uint32 = 0x7f000000 // base group id for custom messages.
)

// TypeIDs
const (
	CommandOKTypeID      uint32 = GroupCommand | TypeIDMaskReply | 0x0000
	CommandErrTypeID     uint32 = GroupCommand | TypeIDMaskReply | 0x0001
	Nav2DCapsQueryTypeID uint32 = GroupNav2D | 0x0000
	Nav2DCapsTypeID      uint32 = Nav2DCapsQueryTypeID | TypeIDMaskReply
	Nav2DDriveTypeID     uint32 = GroupNav2D | 0x0001
	Nav2DTurnTypeID      uint32 = GroupNav2D | 0x0002
)

var (
	// ErrUnknownCommand indicates the command is unknown.
	ErrUnknownCommand = errors.New("unknown command")
)

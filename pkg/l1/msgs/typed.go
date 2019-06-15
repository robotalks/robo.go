package msgs

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"

	fx "github.com/robotalks/robo.go/pkg/framework"
	pb "github.com/robotalks/robo.go/pkg/proto/robo/l1/v1"
)

// TypeID masks
const (
	TypeIDMaskKind  uint32 = 0x80000000
	TypeIDMaskGroup uint32 = 0x7fff0000
	TypeIDMaskID    uint32 = 0x0000ffff
	TypeIDMaskReply uint32 = 0x00008000
)

// Message Kinds
const (
	TypeIDKindCommand uint32 = 0x00000000
	TypeIDKindEvent   uint32 = 0x80000000
)

// Typed wraps a message with type information.
type Typed struct {
	pb.Typed
}

// TypedMsgHandler handles a command-kind message.
type TypedMsgHandler interface {
	HandleTypedMsg(context.Context, fx.Message, *Typed) error
}

// HandleTypedMsgFunc is func form of TypedMsgHandler.
type HandleTypedMsgFunc func(context.Context, fx.Message, *Typed) error

// HandleTypedMsg implements TypedMsgHandler.
func (f HandleTypedMsgFunc) HandleTypedMsg(ctx context.Context, msg fx.Message, typed *Typed) error {
	return f(ctx, msg, typed)
}

// ErrUnknownType indicates unknown type id.
type ErrUnknownType struct {
	TypeID uint32
}

// Error implements error.
func (e *ErrUnknownType) Error() string {
	return fmt.Sprintf("unknown type: %x", e.TypeID)
}

var (
	// ErrNotSerializable indicates the message is not serializable.
	ErrNotSerializable = errors.New("not serializable message")
	// ErrUnsupportedCommand indicates the command is unsupported.
	ErrUnsupportedCommand = errors.New("unsupported command")
)

// SerializableMessage can be serialized over the wire.
type SerializableMessage interface {
	fx.Message
	TypeID() uint32
	Serializable() proto.Message
}

// MessageTypes are predefined mapping of type ID to messages.
var MessageTypes = map[uint32]SerializableMessage{
	CommandOKTypeID:      (*CommandOK)(nil),
	CommandErrTypeID:     (*CommandErr)(nil),
	Nav2DCapsQueryTypeID: (*Nav2DCapsQuery)(nil),
	Nav2DCapsTypeID:      (*Nav2DCaps)(nil),
	Nav2DDriveTypeID:     (*Nav2DDrive)(nil),
	Nav2DTurnTypeID:      (*Nav2DTurn)(nil),
}

// TypedFrom creates a Typed from a serializable message.
func TypedFrom(msg fx.Message) (*Typed, error) {
	if s, ok := msg.(SerializableMessage); ok {
		typeID, serializable := s.TypeID(), s.Serializable()
		data, err := proto.Marshal(serializable)
		if err != nil {
			return nil, err
		}
		return &Typed{Typed: pb.Typed{TypeId: typeID, Message: data}}, nil
	}
	return nil, ErrNotSerializable
}

// Decode decodes the packet into actual message.
func (p Typed) Decode() (fx.Message, error) {
	msgType, ok := MessageTypes[p.TypeId]
	if !ok {
		return nil, &ErrUnknownType{TypeID: p.TypeId}
	}
	msg := msgType.NewMessage()
	serializable := msg.(SerializableMessage).Serializable()
	if err := proto.Unmarshal(p.Message, serializable); err != nil {
		return nil, err
	}
	return msg, nil
}

// Encode encodes the Typed to bytes.
func (p Typed) Encode() ([]byte, error) {
	return proto.Marshal(&p.Typed)
}

// Kind gets message kind from type ID.
func (p Typed) Kind() uint32 {
	return p.TypeId & TypeIDMaskKind
}

// IsCommand determines if the message is a command.
func (p Typed) IsCommand() bool {
	return p.Kind() == TypeIDKindCommand
}

// IsEvent determines if the message is an event.
func (p Typed) IsEvent() bool {
	return p.Kind() == TypeIDKindEvent
}

// DecodeTyped decodes bytes into Typed.
func DecodeTyped(data []byte) (*Typed, error) {
	var typed Typed
	if err := proto.Unmarshal(data, &typed); err != nil {
		return nil, err
	}
	return &typed, nil
}

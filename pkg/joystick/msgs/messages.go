package msgs

import (
	"github.com/golang/protobuf/proto"

	fx "github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/l1/msgs"
)

// JoystickStatusQuery queries the status.
type JoystickStatusQuery struct {
}

// NewMessage implements Message.
func (m *JoystickStatusQuery) NewMessage() fx.Message { return &JoystickStatusQuery{} }

// TypeID implements SerializableMessage.
func (m *JoystickStatusQuery) TypeID() uint32 { return JoystickStatusQueryTypeID }

// Serializable implements SerializableMessage.
func (m *JoystickStatusQuery) Serializable() proto.Message { return m }

// ProtoMessage implements proto.Message.
func (m *JoystickStatusQuery) ProtoMessage() {}

// Reset implements proto.Message.
func (m *JoystickStatusQuery) Reset() { *m = JoystickStatusQuery{} }

// String implements proto.Message.
func (m *JoystickStatusQuery) String() string { return proto.CompactTextString(m) }

// JoystickStatusReply is the response for JoystickStatusQuery.
type JoystickStatusReply struct {
	Status *JoystickStatus `protobuf:"bytes,1,name=status,proto3" json:"status,omitempty"`
}

// NewMessage implements Message.
func (m *JoystickStatusReply) NewMessage() fx.Message { return &JoystickStatusReply{} }

// TypeID implements SerializableMessage.
func (m *JoystickStatusReply) TypeID() uint32 { return JoystickStatusReplyTypeID }

// Serializable implements SerializableMessage.
func (m *JoystickStatusReply) Serializable() proto.Message { return m }

// ProtoMessage implements proto.Message.
func (m *JoystickStatusReply) ProtoMessage() {}

// Reset implements proto.Message.
func (m *JoystickStatusReply) Reset() { *m = JoystickStatusReply{} }

// String implements proto.Message.
func (m *JoystickStatusReply) String() string { return proto.CompactTextString(m) }

// JoystickConnect connects a robot.
type JoystickConnect struct {
	RegistryURL string `protobuf:"bytes,1,opt,name=registry_url,proto3" json:"registry_url,omitempty"`
	Type        string `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	ID          string `protobuf:"bytes,3,opt,name=id,proto3" json:"id,omitempty"`
}

// NewMessage implements Message.
func (m *JoystickConnect) NewMessage() fx.Message { return &JoystickConnect{} }

// TypeID implements SerializableMessage.
func (m *JoystickConnect) TypeID() uint32 { return JoystickConnectTypeID }

// Serializable implements SerializableMessage.
func (m *JoystickConnect) Serializable() proto.Message { return m }

// ProtoMessage implements proto.Message.
func (m *JoystickConnect) ProtoMessage() {}

// Reset implements proto.Message.
func (m *JoystickConnect) Reset() { *m = JoystickConnect{} }

// String implements proto.Message.
func (m *JoystickConnect) String() string { return proto.CompactTextString(m) }

// JoystickStatus is an Event message reflect Joystick status.
type JoystickStatus struct {
	Device     *JoystickDevice  `protobuf:"bytes,1,opt,name=device,proto3" json:"device,omitempty"`
	Connection *JoystickConnect `protobuf:"bytes,2,opt,name=connection,proto3" json:"connection,omitempty"`
}

// NewMessage implements Message.
func (m *JoystickStatus) NewMessage() fx.Message { return &JoystickStatus{} }

// TypeID implements SerializableMessage.
func (m *JoystickStatus) TypeID() uint32 { return JoystickStatusEventTypeID }

// Serializable implements SerializableMessage.
func (m *JoystickStatus) Serializable() proto.Message { return m }

// ProtoMessage implements proto.Message.
func (m *JoystickStatus) ProtoMessage() {}

// Reset implements proto.Message.
func (m *JoystickStatus) Reset() { *m = JoystickStatus{} }

// String implements proto.Message.
func (m *JoystickStatus) String() string { return proto.CompactTextString(m) }

// JoystickDevice provides information of joystick device.
type JoystickDevice struct {
	Index uint32 `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

// GroupJoystick defines the custom group.
const GroupJoystick = msgs.GroupCustom

// TypeIDs
const (
	JoystickStatusEventTypeID uint32 = GroupJoystick | msgs.TypeIDKindEvent | 0x0000
	JoystickStatusQueryTypeID uint32 = GroupJoystick | 0x0000
	JoystickStatusReplyTypeID uint32 = GroupJoystick | msgs.TypeIDMaskReply | 0x0000
	JoystickConnectTypeID     uint32 = GroupJoystick | 0x0001
)

func init() {
	msgs.MessageTypes[JoystickStatusEventTypeID] = (*JoystickStatus)(nil)
	msgs.MessageTypes[JoystickStatusQueryTypeID] = (*JoystickStatusQuery)(nil)
	msgs.MessageTypes[JoystickStatusReplyTypeID] = (*JoystickStatusReply)(nil)
	msgs.MessageTypes[JoystickConnectTypeID] = (*JoystickConnect)(nil)
}

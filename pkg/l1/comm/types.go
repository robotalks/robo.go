package comm

// PacketReader reads packets in bytes.
type PacketReader interface {
	ReadPacket() ([]byte, error)
}

// PacketWriter writes packets in bytes.
type PacketWriter interface {
	WritePacket([]byte) error
}

// PacketReadWriter reads/writes packets in bytes.
type PacketReadWriter interface {
	PacketReader
	PacketWriter
}

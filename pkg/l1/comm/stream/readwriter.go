package stream

import (
	"encoding/binary"
	"io"
)

// ReadWriter implements PacketReadWriter.
// Each packet is prefixed by 4-byte (little-endian) indicate the length.
type ReadWriter struct {
	io.ReadWriter
}

// New creates a ReadWriter with io.ReadWriter.
func New(s io.ReadWriter) *ReadWriter {
	return &ReadWriter{s}
}

// ReadPacket implements PacketReader.
func (p *ReadWriter) ReadPacket() ([]byte, error) {
	var size uint32
	if err := binary.Read(p, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	pkt := make([]byte, size)
	_, err := io.ReadFull(p, pkt)
	return pkt, err
}

// WritePacket implements PacketWriter.
func (p *ReadWriter) WritePacket(pkt []byte) error {
	size := uint32(len(pkt))
	if err := binary.Write(p, binary.LittleEndian, size); err != nil {
		return err
	}
	_, err := p.Write(pkt[:size])
	return err
}

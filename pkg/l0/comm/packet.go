package comm

import (
	"io"
	"time"
)

// PacketSeq defines the type of packet sequence number.
type PacketSeq byte

// NewPacketSeq creates a randome packet sequence number.
func NewPacketSeq() PacketSeq {
	return PacketSeq(byte(time.Now().UnixNano())).Next()
}

// Next calculates the next sequence number.
func (s PacketSeq) Next() PacketSeq {
	n := byte(s) + 1
	if n == 0 || n >= 0xf0 {
		n = 1
	}
	return PacketSeq(n)
}

// IsValid checks if it's a valid sequence number.
func (s PacketSeq) IsValid() bool {
	n := byte(s)
	return n > 0 && n < 0xf0
}

// Packet contains the information of a parsed packet.
type Packet struct {
	Seq  PacketSeq
	Code byte
	Data []byte
}

// Bytes returns encoded bytes for sending.
func (p *Packet) Bytes() []byte {
	b := make([]byte, len(p.Data)+3)
	b[0], b[1] = byte(p.Seq), (p.Code & 0x8f)
	if l := byte(len(p.Data)); l >= 7 {
		b[1] |= 0x70
		b[2] = l
		copy(b[3:], p.Data)
	} else {
		b = b[:l+2]
		b[1] |= (l << 4) & 0x70
		copy(b[2:], p.Data)
	}
	return b
}

// WriteTo writes encoded bytes.
func (p *Packet) WriteTo(w io.Writer) (n int, err error) {
	head := []byte{byte(p.Seq), p.Code & 0x8f, byte(len(p.Data))}
	if head[2] < 7 {
		head[1] |= (head[2] << 4) & 0x70
		head = head[:2]
	} else {
		head[1] |= 0x70
	}
	if n, err = w.Write(head); err != nil {
		return
	}
	if l := byte(len(p.Data)); l > 0 {
		var n1 int
		n1, err = w.Write(p.Data[:l])
		n += n1
	}
	return
}

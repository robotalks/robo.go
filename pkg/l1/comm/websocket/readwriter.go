package websocket

import "golang.org/x/net/websocket"

// ReadWriter implements PacketReadWriter.
type ReadWriter websocket.Conn

// New wraps websocket.Conn.
func New(conn *websocket.Conn) *ReadWriter {
	return (*ReadWriter)(conn)
}

// ReadPacket implements PacketReader.
func (p *ReadWriter) ReadPacket() (pkt []byte, err error) {
	err = websocket.Message.Receive((*websocket.Conn)(p), &pkt)
	return
}

// WritePacket implements PacketWriter.
func (p *ReadWriter) WritePacket(pkt []byte) error {
	return websocket.Message.Send((*websocket.Conn)(p), pkt)
}

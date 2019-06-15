package mqtt

import (
	"context"
	"io"

	"github.com/robotalks/robo.go/pkg/l1"
)

// ReadWriter implements PacketReadWriter.
type ReadWriter struct {
	Queue    *Queue
	SubTopic string
	PubTopic string

	packetCh chan []byte
}

// NewPacketReadWriter creates the ReadWriter.
func NewPacketReadWriter(q *Queue) *ReadWriter {
	return &ReadWriter{Queue: q, packetCh: make(chan []byte, 1)}
}

// WithTopics specifies the topics.
func (p *ReadWriter) WithTopics(sub, pub string) *ReadWriter {
	p.SubTopic, p.PubTopic = sub, pub
	return p
}

// ForConnector sets topics using default convention for connector:
// SubTopic = prefix/msg
// PubTopic = prefix/cmd
func (p *ReadWriter) ForConnector(ref l1.ControllerRef) *ReadWriter {
	prefix := ref.Name()
	return p.WithTopics(prefix+"/msg", prefix+"/cmd")
}

// ForController sets topics using default convention for L1 controller:
// SubTopic = prefix/cmd
// PubTopic = prefix/msg
func (p *ReadWriter) ForController(ref l1.ControllerRef) *ReadWriter {
	prefix := ref.Name()
	return p.WithTopics(prefix+"/cmd", prefix+"/msg")
}

// ReadPacket implements PacketReader.
func (p *ReadWriter) ReadPacket() ([]byte, error) {
	pkt, ok := <-p.packetCh
	if !ok {
		return nil, io.EOF
	}
	return pkt, nil
}

// WritePacket implements PacketWriter.
func (p *ReadWriter) WritePacket(pkt []byte) error {
	token := p.Queue.Pub(p.PubTopic, pkt)
	token.Wait()
	return token.Error()
}

// Run implements Runnable.
func (p *ReadWriter) Run(ctx context.Context) error {
	sub := p.Queue.Sub(p.SubTopic, Handler(p.handleMsg))
	defer sub.Close()
	defer close(p.packetCh)
	<-ctx.Done()
	return ctx.Err()
}

func (p *ReadWriter) handleMsg(_ string, payload []byte) {
	p.packetCh <- payload
}

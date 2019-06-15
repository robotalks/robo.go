package comm

import (
	"container/list"
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testStream struct {
	t          *testing.T
	byteCh     chan byte
	writeCh    chan byte
	injectCh   chan struct{}
	injectList list.List
	injectLock sync.Mutex
}

func newTestStream(t *testing.T) *testStream {
	return &testStream{
		t:        t,
		byteCh:   make(chan byte),
		writeCh:  make(chan byte, 16),
		injectCh: make(chan struct{}, 1),
	}
}

func (s *testStream) Read(p []byte) (int, error) {
	require.Len(s.t, p, 1)
	b, ok := <-s.byteCh
	if ok {
		p[0] = b
		return 1, nil
	}
	return 0, io.EOF
}

func (s *testStream) Write(p []byte) (int, error) {
	for _, b := range p {
		s.writeCh <- b
	}
	return len(p), nil
}

func (s *testStream) run() {
	for {
		var elm *list.Element
		s.injectLock.Lock()
		if s.injectList.Len() > 0 {
			elm = s.injectList.Front()
			s.injectList.Remove(elm)
		}
		s.injectLock.Unlock()
		if elm != nil {
			for _, b := range elm.Value.([]byte) {
				s.byteCh <- b
			}
			continue
		}
		if _, ok := <-s.injectCh; !ok {
			break
		}
	}
}

func (s *testStream) inject(p []byte) {
	if len(p) == 0 {
		return
	}
	s.injectLock.Lock()
	s.injectList.PushBack(p)
	s.injectLock.Unlock()
	select {
	case s.injectCh <- struct{}{}:
	default:
	}
}

type fifoTestCtx struct {
	t            *testing.T
	stream       *testStream
	fifo         *FIFO
	packetCh     chan *Packet
	stateCh      chan SyncState
	expectSeq    PacketSeq
	stateChanges []SyncState
	lock         sync.Mutex
}

func (c *fifoTestCtx) expectStateChanges(expected ...SyncState) *fifoTestCtx {
	if len(expected) > 0 {
		select {
		case <-c.stateCh:
		case <-time.After(500 * time.Millisecond):
			c.t.Fatal("expect state change timeout")
		}
	}
	c.lock.Lock()
	changes := c.stateChanges
	c.stateChanges = nil
	c.lock.Unlock()
	require.Equal(c.t, expected, changes)
	return c
}

func (c *fifoTestCtx) fromPacketSeq(seq PacketSeq) *fifoTestCtx {
	c.expectSeq = seq
	return c
}

func (c *fifoTestCtx) expectPacket(code byte, data []byte) *fifoTestCtx {
	pkt := <-c.packetCh
	require.Equal(c.t, c.expectSeq, pkt.Seq)
	require.Equal(c.t, code, pkt.Code)
	if len(data) > 0 {
		require.Equal(c.t, data, pkt.Data)
	} else {
		require.Empty(c.t, pkt.Data)
	}
	c.expectSeq = c.expectSeq.Next()
	return c
}

func (c *fifoTestCtx) mustSend(code byte, data []byte) *fifoTestCtx {
	err := c.fifo.Send(&Packet{Code: code, Data: data})
	require.NoError(c.t, err)
	return c
}

type fifoTestSequence struct {
	inject []byte
	expect []byte
	action func(int, *fifoTestCtx)
}

type fifoTestCase struct {
	name      string
	sequences []fifoTestSequence
}

func (tc *fifoTestCase) run(t *testing.T) {
	tctx := &fifoTestCtx{
		t:        t,
		stream:   newTestStream(t),
		packetCh: make(chan *Packet, 1),
		stateCh:  make(chan SyncState, 1),
	}
	tctx.fifo = NewFIFO(tctx.stream)
	tctx.fifo.seq = PacketSeq(1)
	tctx.fifo.Handler = HandlePacketFunc(func(ctx context.Context, pkt *Packet) {
		tctx.packetCh <- pkt
	})
	tctx.fifo.Notifier = StateChangedFunc(func(ctx context.Context, state SyncState) {
		tctx.lock.Lock()
		tctx.stateChanges = append(tctx.stateChanges, state)
		tctx.lock.Unlock()
		select {
		case tctx.stateCh <- state:
		default:
		}
	})

	go tctx.stream.run()
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	defer func() { close(tctx.stream.injectCh) }()
	for n, sequence := range tc.sequences {
		tctx.stream.inject(sequence.inject)
		if n == 0 {
			go func() {
				errCh <- tctx.fifo.Run(ctx)
			}()
		}
		for writeLen := 0; writeLen < len(sequence.expect); writeLen++ {
			select {
			case b := <-tctx.stream.writeCh:
				require.Equalf(t, sequence.expect[writeLen], b, "sequences[%d].expect[%d] mismatch", n, writeLen)
			case <-time.After(500 * time.Millisecond):
				t.Fatalf("sequence[%d].expect[%d] timeout", n, writeLen)
			}
		}
		select {
		case err := <-errCh:
			require.NoError(t, err, "FIFO stopped")
		default:
			if a := sequence.action; a != nil {
				a(n, tctx)
			}
		}
	}
}

func TestSync(t *testing.T) {
	cases := []fifoTestCase{
		{
			name: "sync and receive",
			sequences: []fifoTestSequence{
				{
					expect: []byte{syncREQ, 0x01},
				},
				{
					inject: []byte{syncACK, 0x01},
					action: func(n int, tctx *fifoTestCtx) {
						tctx.expectStateChanges(SyncStateReceiving, SyncStateReady)
					},
				},
				{
					inject: []byte{
						0x01, 0x02,
						0x02, 0x92, 0x03,
						0x03, 0x72, 0x08, 1, 2, 3, 4, 5, 6, 7, 8,
					},
					action: func(n int, tctx *fifoTestCtx) {
						tctx.fromPacketSeq(PacketSeq(0x01)).
							expectPacket(0x02, nil).
							expectPacket(0x82, []byte{0x03}).
							expectPacket(0x02, []byte{1, 2, 3, 4, 5, 6, 7, 8})
					},
				},
			},
		},
		{
			name: "sync and send",
			sequences: []fifoTestSequence{
				{
					expect: []byte{syncREQ, 0x01},
				},
				{
					inject: []byte{syncACK, 0x01},
					action: func(n int, tctx *fifoTestCtx) {
						tctx.expectStateChanges(SyncStateReceiving, SyncStateReady).
							mustSend(0x02, nil).
							mustSend(0x82, []byte{0x03}).
							mustSend(0x02, []byte{1, 2, 3, 4, 5, 6, 7, 8})
					},
				},
				{
					expect: []byte{
						0x01, 0x02,
						0x02, 0x92, 0x03,
						0x03, 0x72, 0x08, 1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.run)
	}
}

package comm

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type chanReadWriter struct {
	readCh  <-chan byte
	writeCh chan byte
}

func (c *chanReadWriter) Read(p []byte) (int, error) {
	p[0] = <-c.readCh
	return 1, nil
}

func (c *chanReadWriter) Write(p []byte) (int, error) {
	for _, b := range p {
		c.writeCh <- b
	}
	return len(p), nil
}

type clientTestEnv struct {
	t        *testing.T
	readCh   chan byte
	writeCh  chan byte
	client   *Client
	commands []*Command
}

func newClientTestEnv(t *testing.T) *clientTestEnv {
	env := &clientTestEnv{
		t:       t,
		readCh:  make(chan byte, 1),
		writeCh: make(chan byte, 1),
	}
	clientFIFO := NewFIFO(&chanReadWriter{readCh: env.readCh, writeCh: env.writeCh})
	clientFIFO.seq = PacketSeq(1)
	clientFIFO.ReadTimeout = true
	env.client = NewClient(clientFIFO)
	return env
}

func (e *clientTestEnv) wrapFn(name string, fn func(string)) {
	e.t.Logf("START %s", name)
	fn(name)
	e.t.Logf("STOP %s", name)
}

func (e *clientTestEnv) run(fns ...func(string)) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go e.client.Run(ctx)
	for n, fn := range fns {
		e.wrapFn(fmt.Sprintf("step-%d", n), fn)
	}
}

func (e *clientTestEnv) sequential(fns ...func(string)) func(string) {
	return func(name string) {
		for n, fn := range fns {
			e.wrapFn(name+fmt.Sprintf(".%d", n), fn)
		}
	}
}

func (e *clientTestEnv) parallel(fns ...func(string)) func(string) {
	return func(name string) {
		var wg sync.WaitGroup
		for n, fn := range fns {
			wg.Add(1)
			go func(name string, fn func(string)) {
				defer wg.Done()
				e.wrapFn(name, fn)
			}(name+fmt.Sprintf(".%d", n), fn)
		}
		wg.Wait()
	}
}

func (e *clientTestEnv) expect(bs ...byte) func(string) {
	return func(name string) {
		for i, b := range bs {
			require.Equalf(e.t, b, <-e.writeCh, "%s.byte[%d] mismatch", name, i)
		}
	}
}

func (e *clientTestEnv) inject(bs ...byte) func(string) {
	return func(name string) {
		for _, b := range bs {
			e.readCh <- b
		}
	}
}

func (e *clientTestEnv) stateChange(states ...SyncState) func(string) {
	return func(name string) {
		for i, state := range states {
			require.Equalf(e.t, state, <-e.client.StateChan(), "%s.state[%d] mismatch", name, i)
		}
	}
}

func (e *clientTestEnv) clientDo(code byte, data ...byte) func(string) {
	return func(name string) {
		e.commands = append(e.commands, e.client.Do(&Packet{Code: code, Data: data}))
	}
}

func (e *clientTestEnv) nextResult(name string) (r Result) {
	require.NotEmptyf(e.t, e.commands, "%s commands empty", name)
	cmd := e.commands[0]
	e.commands = e.commands[1:]
	select {
	case r = <-cmd.ResultChan():
	case <-time.After(500 * time.Millisecond):
		e.t.Fatalf("%s: timeout", name)
	}
	return
}

func (e *clientTestEnv) clientResult(code byte, data ...byte) func(string) {
	return func(name string) {
		r := e.nextResult(name)
		require.NoErrorf(e.t, r.Err, "%s unexpected err", name)
		require.Equalf(e.t, code, r.Code, "%s code mismatch", name)
		if len(data) == 0 {
			require.Emptyf(e.t, r.Data, "%s data not empty", name)
		} else {
			require.Equalf(e.t, data, r.Data, "%s data mismatch", name)
		}
	}
}

func (e *clientTestEnv) clientResultErr(err error) func(string) {
	return func(name string) {
		r := e.nextResult(name)
		require.Equalf(e.t, err, r.Err, "%s mismatch", name)
	}
}

func (e *clientTestEnv) clientEvent(code byte, data ...byte) func(string) {
	return func(name string) {
		select {
		case pkt := <-e.client.EventChan():
			require.Equalf(e.t, code, pkt.Code, "%s code mismatch", name)
			if len(data) == 0 {
				require.Emptyf(e.t, pkt.Data, "%s data not empty", name)
			} else {
				require.Equalf(e.t, data, pkt.Data, "%s data mismatch", name)
			}
		case <-time.After(500 * time.Millisecond):
			e.t.Fatalf("%s timeout", name)
		}
	}
}

func TestClient(t *testing.T) {
	testCases := []struct {
		name  string
		logic func(*clientTestEnv)
	}{
		{
			"simple command",
			func(env *clientTestEnv) {
				env.run(
					env.expect(syncREQ, 1),
					env.parallel(
						env.inject(syncACK, 1),
						env.stateChange(SyncStateReceiving, SyncStateReady),
					),
					env.parallel(
						env.clientDo(1),
						env.expect(1, 1),
					),
					env.parallel(
						env.inject(1, 0x10, 1),
						env.stateChange(SyncStateReady|SyncStateReceiving, SyncStateReady),
						env.clientResult(0),
					),
				)
			},
		},
		{
			"no reply",
			func(env *clientTestEnv) {
				env.run(
					env.expect(syncREQ, 1),
					env.parallel(
						env.inject(syncACK, 1),
						env.stateChange(SyncStateReceiving, SyncStateReady),
					),
					env.parallel(
						env.sequential(
							env.clientDo(1),
							env.clientDo(2),
						),
						env.expect(1, 1, 2, 2),
					),
					env.parallel(
						env.inject(1, 0x22, 2, 3),
						env.stateChange(SyncStateReady|SyncStateReceiving, SyncStateReady),
					),
					env.clientResultErr(ErrNoReply),
					env.clientResult(2, 3),
				)
			},
		},
		{
			"event",
			func(env *clientTestEnv) {
				env.run(
					env.expect(syncREQ, 1),
					env.parallel(
						env.inject(syncACK, 1),
						env.stateChange(SyncStateReceiving, SyncStateReady),
					),
					env.parallel(
						env.inject(1, 0x91, 2),
						env.stateChange(SyncStateReady|SyncStateReceiving, SyncStateReady),
					),
					env.clientEvent(0x81, 2),
				)
			},
		},
		{
			"event and command",
			func(env *clientTestEnv) {
				env.run(
					env.expect(syncREQ, 1),
					env.parallel(
						env.inject(syncACK, 1),
						env.stateChange(SyncStateReceiving, SyncStateReady),
					),
					env.parallel(
						env.clientDo(1),
						env.expect(1, 1),
					),
					env.parallel(
						env.inject(1, 0x91, 2),
						env.stateChange(SyncStateReady|SyncStateReceiving, SyncStateReady),
					),
					env.clientEvent(0x81, 2),
					env.parallel(
						env.inject(2, 0x14, 1),
						env.stateChange(SyncStateReady|SyncStateReceiving, SyncStateReady),
					),
					env.clientResult(4),
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := newClientTestEnv(t)
			tc.logic(env)
		})
	}
}

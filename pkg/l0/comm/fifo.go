package comm

import (
	"context"
	"io"
	"os"
	"sync"
	"time"
)

// PacketHandler is called when a packet is received.
type PacketHandler interface {
	HandlePacket(context.Context, *Packet)
}

// HandlePacketFunc is func type of PacketHandler.
type HandlePacketFunc func(context.Context, *Packet)

// HandlePacket implements PacketHandler.
func (f HandlePacketFunc) HandlePacket(ctx context.Context, pkt *Packet) {
	f(ctx, pkt)
}

// StateNotifier is called when packet stream state changed.
type StateNotifier interface {
	StateChanged(context.Context, SyncState)
}

// StateChangedFunc is func type of StateNotifier.
type StateChangedFunc func(context.Context, SyncState)

// StateChanged implements StateNotifier.
func (f StateChangedFunc) StateChanged(ctx context.Context, state SyncState) {
	f(ctx, state)
}

// FIFO send/recv packets.
type FIFO struct {
	ReadWriter  io.ReadWriter
	Handler     PacketHandler
	Notifier    StateNotifier
	Timeout     time.Duration
	ReadTimeout bool // set to true if ReadWriter already supports timeout with Read

	seq   PacketSeq
	state SyncState
	lock  sync.RWMutex

	syncTimer <-chan time.Time
	parser    Parser
}

// NewFIFO creates a FIFO.
func NewFIFO(rw io.ReadWriter) *FIFO {
	return &FIFO{
		ReadWriter: rw,
		Timeout:    100 * time.Millisecond,
		seq:        NewPacketSeq(),
	}
}

// State gets the state.
func (f *FIFO) State() SyncState {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.state
}

// Send sends a packet.
func (f *FIFO) Send(pkt *Packet) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	if !f.state.IsReady() {
		return ErrNotReady
	}
	pkt.Seq = f.seq
	if _, err := pkt.WriteTo(f.ReadWriter); err != nil {
		return err
	}
	f.seq = f.seq.Next()
	return nil
}

// Run processes the FIFO in the background.
func (f *FIFO) Run(ctx context.Context) error {
	err := f.applyParseResult(ctx, f.parser.Reset())
	if err != nil {
		return err
	}

	if f.ReadTimeout {
		buf := make([]byte, 1)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-f.syncTimer:
				if err = f.applyParseResult(ctx, f.parser.Timeout()); err != nil {
					return err
				}
			default:
				n, err := f.ReadWriter.Read(buf)
				if err != nil {
					if os.IsTimeout(err) {
						err = f.applyParseResult(ctx, f.parser.Timeout())
					}
				} else if n == 0 {
					err = f.applyParseResult(ctx, f.parser.Timeout())
				} else {
					err = f.applyParseResult(ctx, f.parser.Parse(buf[0]))
				}
				if err != nil {
					return err
				}
			}
		}
	} else {
		byteCh, errCh := make(chan byte), make(chan error, 1)
		subCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go f.readLoop(subCtx, byteCh, errCh)
		for {
			select {
			case b := <-byteCh:
				if err = f.applyParseResult(ctx, f.parser.Parse(b)); err != nil {
					return err
				}
			case err := <-errCh:
				return err
			case <-ctx.Done():
				return ctx.Err()
			case <-f.syncTimer:
				if err = f.applyParseResult(ctx, f.parser.Timeout()); err != nil {
					return err
				}
			}
		}
	}
}

func (f *FIFO) readLoop(ctx context.Context, byteCh chan byte, errCh chan error) {
	buf := make([]byte, 1)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := f.ReadWriter.Read(buf)
			if err != nil {
				errCh <- err
				return
			}
			byteCh <- buf[0]
		}
	}
}

func (f *FIFO) applyParseResult(ctx context.Context, pr ParseResult) (err error) {
	var notifier StateNotifier
	f.lock.Lock()
	if f.state != pr.State {
		f.state = pr.State
		notifier = f.Notifier
	}
	if pr.Sync != 0 {
		_, err = f.ReadWriter.Write([]byte{pr.Sync, byte(f.seq)})
	}
	f.lock.Unlock()
	if err != nil {
		return
	}

	if f.ReadTimeout {
		if pr.Sync == syncREQ {
			f.syncTimer = time.After(f.Timeout)
		} else {
			f.syncTimer = nil
		}
	} else {
		switch pr.WhatAboutTimer() {
		case TimerRestart:
			f.syncTimer = time.After(f.Timeout)
		case TimerStop:
			f.syncTimer = nil
		}
	}

	if notifier != nil {
		notifier.StateChanged(ctx, pr.State)
	}
	if pr.Packet != nil {
		if h := f.Handler; h != nil {
			h.HandlePacket(ctx, pr.Packet)
		}
	}
	return
}

package comm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type parserTestSequence struct {
	in     []byte
	expect ParseResult
	final  ParseResult
}

type parserTestSequenceBuilder struct {
	seq []parserTestSequence
}

func parserTestSequences() *parserTestSequenceBuilder {
	return &parserTestSequenceBuilder{}
}

func (b *parserTestSequenceBuilder) on(state SyncState, in ...byte) *parserTestSequenceBuilder {
	s := parserTestSequence{in: in, expect: ParseResult{State: state}}
	s.final = s.expect
	b.seq = append(b.seq, s)
	return b
}

func (b *parserTestSequenceBuilder) onSyncing(in ...byte) *parserTestSequenceBuilder {
	return b.on(SyncStateSyncing|SyncStateReceiving, in...)
}

func (b *parserTestSequenceBuilder) onReceiving(in ...byte) *parserTestSequenceBuilder {
	return b.on(SyncStateReady|SyncStateReceiving, in...)
}

func (b *parserTestSequenceBuilder) timeout() *parserTestSequenceBuilder {
	b.seq = append(b.seq, parserTestSequence{})
	return b
}

func (b *parserTestSequenceBuilder) expect(pr ParseResult) *parserTestSequenceBuilder {
	b.seq[len(b.seq)-1].expect = pr
	return b
}

func (b *parserTestSequenceBuilder) final(pr ParseResult) *parserTestSequenceBuilder {
	b.seq[len(b.seq)-1].final = pr
	return b
}

func (b *parserTestSequenceBuilder) synced() *parserTestSequenceBuilder {
	return b.final(ParseResult{State: SyncStateReady})
}

func (b *parserTestSequenceBuilder) packet(seq, code byte, data ...byte) *parserTestSequenceBuilder {
	return b.final(ParseResult{State: SyncStateReady, Packet: &Packet{Seq: PacketSeq(seq), Code: code, Data: data}})
}

func (b *parserTestSequenceBuilder) resync() *parserTestSequenceBuilder {
	return b.final(ParseResult{Sync: syncREQ, State: SyncStateSyncing})
}

func (b *parserTestSequenceBuilder) syncedWithAck() *parserTestSequenceBuilder {
	return b.final(ParseResult{Sync: syncACK, State: SyncStateReady})
}

func (b *parserTestSequenceBuilder) build() []parserTestSequence {
	return b.seq
}

func TestParser(t *testing.T) {
	testCases := []struct {
		name string
		seq  []parserTestSequence
	}{
		{
			name: "sync and receive",
			seq: parserTestSequences().
				onSyncing(syncACK, 1).synced().
				onReceiving(1, 0x02).packet(1, 2).
				onReceiving(2, 0x72, 0).packet(2, 2).
				onReceiving(3, 0x92, 0x03).packet(3, 0x82, 3).
				onReceiving(4, 0x72, 0x08, 1, 2, 3, 4, 5, 6, 7, 8).packet(4, 2, 1, 2, 3, 4, 5, 6, 7, 8).
				build(),
		},
		{
			name: "sync timeout",
			seq: parserTestSequences().
				timeout().resync().
				onSyncing(syncACK).
				timeout().resync().
				build(),
		},
		{
			name: "sync skip invalid bytes",
			seq: parserTestSequences().
				on(SyncStateSyncing, 1, 2, 3, 4, 0x80, 0x81, 0xf0, 0xf1).
				onSyncing(syncACK, 1).synced().
				build(),
		},
		{
			name: "handle req in sync",
			seq: parserTestSequences().
				onSyncing(syncREQ, 1).syncedWithAck().
				build(),
		},
		{
			name: "handle req in sync with invalid seq",
			seq: parserTestSequences().
				onSyncing(syncREQ, syncREQ).resync().
				onSyncing(syncACK, 1).synced().
				build(),
		},
		{
			name: "handle req after sync",
			seq: parserTestSequences().
				onSyncing(syncACK, 1).synced().
				onSyncing(syncREQ, 1).syncedWithAck().
				onReceiving(1, 0x02).packet(1, 2).
				build(),
		},
		{
			name: "handle req after sync with invalid seq",
			seq: parserTestSequences().
				onSyncing(syncACK, 1).synced().
				onSyncing(syncREQ, syncACK).resync().
				onSyncing(syncACK, 1).synced().
				build(),
		},
		{
			name: "handle ack in sync with invalid seq",
			seq: parserTestSequences().
				onSyncing(syncACK, syncREQ).resync().
				onSyncing(syncACK, 1).synced().
				build(),
		},
		{
			name: "handle ack after sync",
			seq: parserTestSequences().
				onSyncing(syncACK, 1).synced().
				onReceiving(syncACK, 1).synced().
				onReceiving(1, 0x02).packet(1, 2).
				build(),
		},
		{
			name: "ack invalid seq after sync",
			seq: parserTestSequences().
				onSyncing(syncACK, 1).synced().
				onReceiving(syncACK, 2).resync().
				onSyncing(syncACK, 2).synced().
				onReceiving(2, 0x02).packet(2, 2).
				build(),
		},
		{
			name: "invalid seq",
			seq: parserTestSequences().
				onSyncing(syncACK, 1).synced().
				onReceiving(1, 2).packet(1, 2).
				onSyncing(1).resync().
				on(SyncStateSyncing, 0x92, 3).
				onSyncing(syncACK, 3).synced().
				build(),
		},
		{
			name: "invalid data len",
			seq: parserTestSequences().
				onSyncing(syncACK, 1).synced().
				onReceiving(1, 0x70, 0x80).resync().
				on(SyncStateSyncing, 1, 2, 3, 4).
				onSyncing(syncACK, 1).synced().
				build(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var parser Parser
			for n, s := range tc.seq {
				var pr ParseResult
				if l := len(s.in); l == 0 {
					pr = parser.Timeout()
				} else {
					for i, b := range s.in {
						pr = parser.Parse(b)
						if i+1 < l {
							require.Equalf(t, s.expect, pr, "seq[%d][%d] expect mismatch", n, i)
						}
					}
				}
				require.Equalf(t, s.final, pr, "seq[%d] final mismatch", n)
			}
		})
	}
}

func TestParserReset(t *testing.T) {
	var parser Parser
	pr := parser.Reset()
	require.Equal(t, syncREQ, pr.Sync)
	require.Equal(t, SyncStateSyncing, pr.State)
	require.Nil(t, pr.Packet)
}

func TestSyncState(t *testing.T) {
	require.False(t, SyncStateSyncing.IsReady())
	require.False(t, SyncStateSyncing.IsReceiving())
	require.True(t, SyncStateReady.IsReady())
	require.False(t, SyncStateReady.IsReceiving())
	require.False(t, SyncStateReceiving.IsReady())
	require.True(t, SyncStateReceiving.IsReceiving())
	require.True(t, (SyncStateReady | SyncStateReceiving).IsReady())
	require.True(t, (SyncStateReady | SyncStateReceiving).IsReceiving())
}

func TestParseResult(t *testing.T) {
	testCases := []struct {
		state  SyncState
		cmd    byte
		action TimerAction
	}{
		{SyncStateSyncing, 0, TimerNoChange},
		{SyncStateSyncing, syncACK, TimerNoChange},
		{SyncStateSyncing, syncREQ, TimerRestart},
		{SyncStateSyncing, syncACK, TimerNoChange},
		{SyncStateReceiving, 0, TimerRestart},
		{SyncStateReady, 0, TimerStop},
		{SyncStateReady, syncACK, TimerStop},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%x %x", tc.state, tc.cmd), func(t *testing.T) {
			pr := ParseResult{Sync: tc.cmd, State: tc.state}
			require.Equal(t, tc.action, pr.WhatAboutTimer())
		})
	}
}

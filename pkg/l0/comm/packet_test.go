package comm

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPacketSeq(t *testing.T) {
	for s := byte(0xff); s >= byte(0xf0); s-- {
		require.False(t, PacketSeq(s).IsValid())
		require.Equal(t, PacketSeq(1), PacketSeq(s).Next())
	}
	for s := byte(1); s < byte(0xf0); s++ {
		require.True(t, PacketSeq(s).IsValid())
		if s+1 < 0xf0 {
			require.Equal(t, PacketSeq(s+1), PacketSeq(s).Next())
		} else {
			require.Equal(t, PacketSeq(1), PacketSeq(s).Next())
		}
	}
	require.False(t, PacketSeq(0).IsValid())
	require.Equal(t, PacketSeq(1), PacketSeq(0).Next())
}

func TestPacket(t *testing.T) {
	testCases := []struct {
		name   string
		packet Packet
		expect []byte
	}{
		{"no data", Packet{Seq: PacketSeq(1), Code: 2}, []byte{1, 2}},
		{"small data", Packet{Seq: PacketSeq(1), Code: 2, Data: []byte{1}}, []byte{1, 0x12, 1}},
		{"large data", Packet{Seq: PacketSeq(1), Code: 2, Data: []byte{1, 2, 3, 4, 5, 6, 7}}, []byte{1, 0x72, 7, 1, 2, 3, 4, 5, 6, 7}},
		{"event no data", Packet{Seq: PacketSeq(1), Code: 0x82}, []byte{1, 0x82}},
		{"event small data", Packet{Seq: PacketSeq(1), Code: 0x82, Data: []byte{1}}, []byte{1, 0x92, 1}},
		{"event large data", Packet{Seq: PacketSeq(1), Code: 0x82, Data: []byte{1, 2, 3, 4, 5, 6, 7}}, []byte{1, 0xf2, 7, 1, 2, 3, 4, 5, 6, 7}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expect, tc.packet.Bytes())
			var buf bytes.Buffer
			n, err := tc.packet.WriteTo(&buf)
			require.NoError(t, err)
			require.Equal(t, tc.expect, buf.Bytes())
			require.Equal(t, len(tc.expect), n)
		})
	}
}

package comm

// Parser parses bytes received.
type Parser struct {
	peerSeq PacketSeq
	state   parseState
	packet  *Packet
	recvLen byte
}

// SyncState indicates the state of communication.
type SyncState int

const (
	// SyncStateSyncing means the communication is not synchronized.
	SyncStateSyncing SyncState = 0
	// SyncStateReady means the communication is synchronized and ready for packets.
	SyncStateReady SyncState = 0x01
	// SyncStateReceiving means there's on-going communication for syncing or a packet.
	SyncStateReceiving SyncState = 0x02
)

// IsReady indicates if the communication is ready for packets.
func (s SyncState) IsReady() bool {
	return s&SyncStateReady != 0
}

// IsReceiving indicates if it's in the middle for syncing or receiving a packet.
func (s SyncState) IsReceiving() bool {
	return s&SyncStateReceiving != 0
}

// TimerAction defines what to do with timer.
type TimerAction int

const (
	// TimerNoChange indicates keep the timer as-is.
	TimerNoChange TimerAction = iota
	// TimerRestart to restart the timer.
	TimerRestart
	// TimerStop to stop/cancel the timer.
	TimerStop
)

// ParseResult indicates the result after one parsing step.
type ParseResult struct {
	Sync   byte
	State  SyncState
	Packet *Packet
}

// WhatAboutTimer decides what to do with timer.
func (r ParseResult) WhatAboutTimer() TimerAction {
	if r.State.IsReceiving() || r.Sync == syncREQ {
		return TimerRestart
	}
	if r.State.IsReady() {
		return TimerStop
	}
	return TimerNoChange
}

type parseState int

const (
	stateSyncAck    parseState = iota // sync req sent, waiting for syncACK
	stateSyncReqSeq                   // waiting for sync seq after syncREQ
	stateSyncAckSeq                   // waiting for sync seq after syncACK
	stateMsgSeq                       // waiting for message seq
	stateMsgAckSeq                    // recv ack in MsgSeq, validate seq
	stateMsgCode                      // waiting for message code
	stateMsgLen                       // waiting for message length
	stateMsgData                      // waiting for message data
)

const (
	syncREQ byte = 0xff
	syncACK byte = 0xfe
)

// State gets the current sync state.
func (p *Parser) State() SyncState {
	if p.state == stateSyncAck {
		return SyncStateSyncing
	}
	if p.state == stateMsgSeq {
		return SyncStateReady
	}
	if p.state > stateMsgSeq {
		return SyncStateReady | SyncStateReceiving
	}
	return SyncStateSyncing | SyncStateReceiving
}

// Reset resets the internal state of parser.
func (p *Parser) Reset() (pr ParseResult) {
	p.packet = nil
	pr.Sync, pr.Packet = p.resync()
	pr.State = p.State()
	return
}

// Parse consumes one byte.
func (p *Parser) Parse(b byte) (pr ParseResult) {
	pr.Sync, pr.Packet = p.parseByte(b)
	pr.State = p.State()
	return
}

// Timeout notifies the parser timer expires.
func (p *Parser) Timeout() (pr ParseResult) {
	if p.state != stateMsgSeq {
		pr.Sync, pr.Packet = p.resync()
	}
	pr.State = p.State()
	return
}

func (p *Parser) parseByte(b byte) (syncCmd byte, pkt *Packet) {
	switch p.state {
	case stateSyncAck:
		switch b {
		case syncREQ:
			p.state = stateSyncReqSeq
		case syncACK:
			p.state = stateSyncAckSeq
		}
	case stateSyncReqSeq:
		if seq := PacketSeq(b); seq.IsValid() {
			p.peerSeq, p.state = seq, stateMsgSeq
			syncCmd = syncACK
			return
		}
		return p.resync()
	case stateSyncAckSeq:
		if seq := PacketSeq(b); seq.IsValid() {
			p.peerSeq, p.state = seq, stateMsgSeq
			return
		}
		return p.resync()
	case stateMsgSeq:
		if b == syncREQ {
			p.state = stateSyncReqSeq
			return
		}
		if b == syncACK {
			p.state = stateMsgAckSeq
			return
		}
		if b != byte(p.peerSeq) {
			return p.resync()
		}
		p.packet = &Packet{Seq: p.peerSeq}
		p.peerSeq = p.peerSeq.Next()
		p.state = stateMsgCode
	case stateMsgAckSeq:
		if b != byte(p.peerSeq) {
			return p.resync()
		}
		p.state = stateMsgSeq
	case stateMsgCode:
		p.packet.Code = b & 0x8f
		switch dataLen := (b >> 4) & 7; dataLen {
		case 0:
			return p.packetReady()
		case 7:
			p.state = stateMsgLen
		default:
			p.packet.Data, p.recvLen = make([]byte, dataLen), 0
			p.state = stateMsgData
		}
	case stateMsgLen:
		if b >= 0x80 {
			return p.resync()
		}
		if b == 0 {
			return p.packetReady()
		}
		p.packet.Data, p.recvLen = make([]byte, b), 0
		p.state = stateMsgData
	case stateMsgData:
		p.packet.Data[p.recvLen] = b
		p.recvLen++
		if p.recvLen >= byte(len(p.packet.Data)) {
			return p.packetReady()
		}
	}
	return
}

func (p *Parser) resync() (byte, *Packet) {
	p.state = stateSyncAck
	return syncREQ, nil
}

func (p *Parser) packetReady() (syncCmd byte, pkt *Packet) {
	p.state = stateMsgSeq
	pkt, p.packet = p.packet, nil
	return
}

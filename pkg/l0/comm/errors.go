package comm

import (
	"errors"
	"fmt"
)

var (
	// ErrNotReady indicates the FIFO is not ready for communication.
	ErrNotReady = errors.New("not ready")
	// ErrNoReply indicates no reply received from peer.
	// This happens when a reply is received for a latter command, and all
	// previous commands fail with this error.
	ErrNoReply = errors.New("no reply")
)

// CommandError wraps error codes from reply.
type CommandError struct {
	Code byte
}

// Error implements error.
func (e *CommandError) Error() string {
	return fmt.Sprintf("command error %d", e.Code)
}

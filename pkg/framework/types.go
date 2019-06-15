package framework

import (
	"context"
	"time"
)

// Named is an abstraction for things with a name.
type Named interface {
	Name() string
}

// Runnable defines a generic interface for background runners.
type Runnable interface {
	Run(context.Context) error
}

// Message defines the abstract message to be
// consumed in a controlling loop.
type Message interface {
	// NewMessage creates an empty message.
	NewMessage() Message
}

// MessageHandler processes a message.
type MessageHandler interface {
	HandleMessage(context.Context, Message)
}

// HandleMessageFunc is the func form of MessageHandler.
type HandleMessageFunc func(context.Context, Message)

// HandleMessage implements MessageHandler.
func (f HandleMessageFunc) HandleMessage(ctx context.Context, msg Message) {
	f(ctx, msg)
}

// Controller defines the abstract controlling logic.
type Controller interface {
	Control(ControlContext) error
}

// TimeSource provides the time for controlling logic.
type TimeSource interface {
	Time() time.Time
}

// ControlContext provides the context of current control
// iteration.
type ControlContext interface {
	TimeSource
	// Context retrieves context.Context.
	Context() context.Context
	// PriorityLevel gets the current priority level.
	PriorityLevel() int
	// Messages retrieves all messages collected when
	// this iteration starts.
	Messages() MessageStore
	// PostRun injects post-run one-shot hooks at current
	// priority level. If called in post-run hooks, new hooks
	// are installed for next iteration.
	PostRun(hooks ...Controller)

	LoopControl
}

// PriorityLevels is the total levels of priorities.
const PriorityLevels int = 16

// Predefine priority levels
const (
	PrLvTop    int = 0
	PrLvHigh   int = 4
	PrLvNormal int = 8
	PrLvLow    int = 12
	PrLvIdle   int = PriorityLevels - 1

	// PrLvSense is the alias of priority level for sensors.
	PrLvSense = PrLvHigh
	// PrLvControl is the alias of priority level for controllers.
	PrLvControl = PrLvNormal
	// PrLvAcuate is the alias of priority level for acuators.
	PrLvAcuate = PrLvLow
	// PrLvPostProc is the alias of priority level for post-processing.
	PrLvPostProc = PrLvIdle - 1
)

// LoopControl exposes access to the controlling loop.
type LoopControl interface {
	// PreRunAt injects one-shot pre-run controller hooks at
	// specified priority level.
	PreRunAt(priorityLevel int, controllers ...Controller)
	// PostRunAt injects one-shot post-run controller hooks at
	// specified priority level.
	PostRunAt(priorityLevel int, controllers ...Controller)
	// PostMessage enqueues the message.
	PostMessage(Message)
	// TriggerNext schedules the next iteration to be executed
	// immediately after the current iteration.
	TriggerNext()
}

// MessageStore provides read/write access to a list of messages.
type MessageStore interface {
	// ProcessMessages uses a processor to process all messages.
	ProcessMessages(MessageProcessor)

	MessageAppender
}

// MessageAppender appends message to store.
type MessageAppender interface {
	// AddMessages appends messages to the store for next processing cycle.
	AddMessages(msgs ...Message)
}

// MessageProcessor is used by MessageStore to process messages.
type MessageProcessor interface {
	ProcessMessage(MessageProcessingContext)
}

// ProcessMessageFunc is the func form of MessageProcessor.
type ProcessMessageFunc func(MessageProcessingContext)

// ProcessMessage implements MessageProcessor.
func (f ProcessMessageFunc) ProcessMessage(mc MessageProcessingContext) {
	f(mc)
}

// MessageProcessingContext provides context for current message.
type MessageProcessingContext interface {
	// CurrentMessage gets the current message being processed.
	CurrentMessage() Message
	// MessageTaken indicates the message has been processed and
	// should be removed from store.
	MessageTaken()
	// StopProcessing indicates no need to examine further messages.
	StopProcessing()

	MessageAppender
}

// ControlFunc defines the func form of Controller.
type ControlFunc func(ControlContext) error

// Control implements Controller.
func (f ControlFunc) Control(ctx ControlContext) error {
	return f(ctx)
}

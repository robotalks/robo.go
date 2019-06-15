package framework

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/golang/glog"
)

// Loop manages sensors, controllers, acuators.
type Loop struct {
	Interval time.Duration

	controllers [PriorityLevels]controllerList

	runners []Runnable

	messages messageList
	lock     sync.Mutex

	wakeUpCh chan struct{}
}

// LoopAdder provides specific logic to add components to loop.
type LoopAdder interface {
	AddToLoop(*Loop)
}

type loopCtl struct {
	*Loop
}

type loopIteration struct {
	loopCtl
	ctx           context.Context
	time          time.Time
	priorityLevel int
	messages      messageList
}

type messageList struct {
	head *messageItem
	tail *messageItem
}

type messageItem struct {
	msg  Message
	next *messageItem
}

func (l *messageList) append(item *messageItem) {
	if l.head == nil {
		l.head = item
	} else {
		l.tail.next = item
	}
	l.tail = item
}

func (l *messageList) splice(src *messageList) {
	l.head, l.tail, src.head = src.head, src.tail, nil
}

func (l *messageList) concat(lst *messageList) {
	if l.head == nil {
		l.head = lst.head
	} else {
		l.tail.next = lst.head
	}
	if lst.head != nil {
		l.tail = lst.tail
	}
}

type controllerList struct {
	preHooks    []Controller
	controllers []Controller
	postHooks   []Controller
	lock        sync.Mutex
}

var (
	loopCtxKey = &Loop{}
)

// LoopCtlFrom gets LoopCtl from context.
func LoopCtlFrom(ctx context.Context) LoopControl {
	return ctx.Value(loopCtxKey).(LoopControl)
}

// CtlCtxFrom gets ControlContext from context.
func CtlCtxFrom(ctx context.Context) ControlContext {
	return ctx.Value(loopCtxKey).(ControlContext)
}

// NewLoop creates a Loop.
func NewLoop() *Loop {
	return &Loop{Interval: 100 * time.Millisecond}
}

// Add adds LoopAdders.
func (l *Loop) Add(adders ...LoopAdder) *Loop {
	for _, adder := range adders {
		adder.AddToLoop(l)
	}
	return l
}

// AddController registers controllers to the loop.
func (l *Loop) AddController(priorityLevel int, ctls ...Controller) *Loop {
	lst := &l.controllers[priorityLevel]
	lst.controllers = append(lst.controllers, ctls...)
	for _, ctl := range ctls {
		if runner, ok := ctl.(Runnable); ok {
			l.runners = append(l.runners, runner)
		}
	}
	return l
}

// AddRunnable adds Runnable implementions.
func (l *Loop) AddRunnable(runnables ...Runnable) *Loop {
	l.runners = append(l.runners, runnables...)
	return l
}

// Run implements Runnable.
func (l *Loop) Run(ctx context.Context) error {
	if l.wakeUpCh == nil {
		l.wakeUpCh = make(chan struct{}, 1)
	}

	runner := NewRunnerWith(context.WithValue(ctx, loopCtxKey, &loopCtl{l}))
	runner.Go(l.runners...)
	defer runner.Wait()

	interval := l.Interval
	if interval == 0 {
		interval = 100 * time.Millisecond
	}
	timer := time.Tick(interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer:
			l.runIteration(ctx)
		case <-l.wakeUpCh:
			l.runIteration(ctx)
		}
	}
}

// RunOrFail is intended to be used in main to simply run the loop.
func (l *Loop) RunOrFail() {
	if err := l.Run(context.TODO()); err != nil {
		log.Fatalln(err)
	}
}

// PreRunAt implements LoopCtl.
func (l *Loop) PreRunAt(priorityLevel int, hooks ...Controller) {
	lst := &l.controllers[priorityLevel]
	lst.lock.Lock()
	lst.preHooks = append(lst.preHooks, hooks...)
	lst.lock.Unlock()
}

// PostRunAt implements LoopCtl.
func (l *Loop) PostRunAt(priorityLevel int, hooks ...Controller) {
	lst := &l.controllers[priorityLevel]
	lst.lock.Lock()
	lst.postHooks = append(lst.postHooks, hooks...)
	lst.lock.Unlock()
}

// PostMessage implements LoopCtl.
func (l *Loop) PostMessage(msg Message) {
	l.lock.Lock()
	l.messages.append(&messageItem{msg: msg})
	l.lock.Unlock()
}

// TriggerNext implements LoopCtl.
func (l *Loop) TriggerNext() {
	select {
	case l.wakeUpCh <- struct{}{}:
	default:
	}
}

func (l *Loop) runIteration(ctx context.Context) {
	iter := &loopIteration{loopCtl: loopCtl{l}, time: time.Now()}
	l.lock.Lock()
	iter.messages.splice(&l.messages)
	l.lock.Unlock()
	iter.ctx = context.WithValue(ctx, loopCtxKey, iter)
	for i := 0; i < PriorityLevels; i++ {
		iter.priorityLevel = i
		l.controllers[i].run(iter)
	}
}

func (t *loopIteration) Context() context.Context {
	return t.ctx
}

func (t *loopIteration) Time() time.Time {
	return t.time
}

func (t *loopIteration) PriorityLevel() int {
	return t.priorityLevel
}

func (t *loopIteration) Messages() MessageStore {
	return t
}

func (t *loopIteration) PostRun(hooks ...Controller) {
	t.PostRunAt(t.priorityLevel, hooks...)
}

// MessageStore implementations

type messageContext struct {
	iter  *loopIteration
	item  *messageItem
	taken bool
	stop  bool
}

func (c *messageContext) CurrentMessage() Message     { return c.item.msg }
func (c *messageContext) MessageTaken()               { c.taken = true }
func (c *messageContext) StopProcessing()             { c.stop = true }
func (c *messageContext) AddMessages(msgs ...Message) { c.iter.AddMessages(msgs...) }

func (t *loopIteration) ProcessMessages(proc MessageProcessor) {
	var msgs, remains messageList
	msgs.splice(&t.messages)
	for msgs.head != nil {
		mctx := &messageContext{iter: t, item: msgs.head}
		msgs.head = msgs.head.next
		mctx.item.next = nil
		proc.ProcessMessage(mctx)
		if !mctx.taken {
			remains.append(mctx.item)
		}
		if mctx.stop {
			remains.concat(&msgs)
		}
	}
	remains.concat(&t.messages)
	t.messages = remains
}

func (t *loopIteration) AddMessages(msgs ...Message) {
	for _, msg := range msgs {
		t.messages.append(&messageItem{msg: msg})
	}
}

func (c *controllerList) run(iter *loopIteration) {
	c.lock.Lock()
	ctls := c.preHooks
	c.preHooks = nil
	c.lock.Unlock()
	runControllers(iter, ctls)
	runControllers(iter, c.controllers)
	c.lock.Lock()
	ctls, c.postHooks = c.postHooks, nil
	c.lock.Unlock()
	runControllers(iter, ctls)
}

func runControllers(iter *loopIteration, ctls []Controller) {
	for _, ctl := range ctls {
		if err := ctl.Control(iter); err != nil {
			glog.Errorf("controller error: %v", err)
		}
	}
}

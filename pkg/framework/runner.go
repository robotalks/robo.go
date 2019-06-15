package framework

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/golang/glog"
)

type namedRunnable struct {
	Runnable
	name string
}

func (r *namedRunnable) Name() string {
	return r.name
}

// NamedRun wraps a Runnable with a name.
func NamedRun(name string, runnable Runnable) Runnable {
	return &namedRunnable{name: name, Runnable: runnable}
}

// Runner runs multiple Runnables and collect errors.
type Runner struct {
	Context context.Context
	Runners []Runnable

	errCh  chan error
	exitCh chan struct{}
}

// NewRunner creates a runner with a default background context.
func NewRunner() *Runner {
	return NewRunnerWith(context.Background())
}

// NewRunnerWith creates a runner with a specified context.
func NewRunnerWith(ctx context.Context) *Runner {
	return &Runner{
		Context: ctx,
		errCh:   make(chan error, 1),
		exitCh:  make(chan struct{}),
	}
}

// HandleSignals handles CtrlC and SIGTERM from the system.
func (r *Runner) HandleSignals() *Runner {
	ctx, cancel := context.WithCancel(r.Context)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	r.Context = ctx
	go func() {
		<-sigCh
		glog.Info("stop requested")
		cancel()
		<-sigCh
		glog.Error("stop requested again, force exit")
		close(r.exitCh)
	}()
	return r
}

// Go spawns a Runnable with default context.
func (r *Runner) Go(runners ...Runnable) *Runner {
	return r.GoWith(r.Context, runners...)
}

// GoWith spawns a Runnable with a specified context.
func (r *Runner) GoWith(ctx context.Context, runners ...Runnable) *Runner {
	for _, runner := range runners {
		var name string
		if named, ok := runner.(Named); ok {
			name = named.Name()
		} else {
			name = strconv.Itoa(len(r.Runners))
		}
		r.Runners = append(r.Runners, runner)
		glog.V(4).Infof("start Runner[%s]", name)
		go func(runner Runnable, name string) {
			glog.V(4).Infof("Runner[%s] started", name)
			r.errCh <- runner.Run(ctx)
			glog.V(4).Infof("Runner[%s] stopped", name)
		}(runner, name)
	}
	return r
}

// Wait waits until all Runnables stops and aggregate errors.
func (r *Runner) Wait() error {
	var errs AggregatedError
	for range r.Runners {
		select {
		case <-r.exitCh:
			return errors.New("forced exit")
		case err := <-r.errCh:
			if err != context.Canceled {
				errs.Add(err)
			}
		}
	}
	return errs.Aggregate()
}

// RunWithContextCancel runs a func with doesn't accept a context.
// cancel is called only when the context is canceled.
func RunWithContextCancel(ctx context.Context, onCancel func(), fn func() error) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- fn()
	}()
	select {
	case <-ctx.Done():
		if onCancel != nil {
			onCancel()
		}
		<-errCh
		return context.Canceled
	case err := <-errCh:
		return err
	}
}

// RunWithContext is simplified form with no cancel callback.
func RunWithContext(ctx context.Context, fn func() error) error {
	return RunWithContextCancel(ctx, nil, fn)
}

// RunWithContextCloser is a convinient wrapper for RunWithContextCancel and
// ensures closer.Close is either called on cancel or exit of fn.
func RunWithContextCloser(ctx context.Context, closer io.Closer, fn func() error) error {
	var closed bool
	err := RunWithContextCancel(ctx, func() {
		closer.Close()
		closed = true
	}, fn)
	if !closed {
		closer.Close()
	}
	return err
}

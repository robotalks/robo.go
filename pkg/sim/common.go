package sim

import (
	fx "github.com/robotalks/robo.go/pkg/framework"
)

// ObjectsChangeCaster provides a subscriber and implements
// listener to cast notifcations.
type ObjectsChangeCaster struct {
	listeners []ObjectsChangeListener
}

// SubscribeObjectsChange implements ObjectsChangeSubscriber.
func (c *ObjectsChangeCaster) SubscribeObjectsChange(ln ObjectsChangeListener) {
	c.listeners = append(c.listeners, ln)
}

// ObjectsChanged implements ObjectsChangeListener.
func (c *ObjectsChangeCaster) ObjectsChanged(cc fx.ControlContext, objs ...Object) {
	for _, ln := range c.listeners {
		ln.ObjectsChanged(cc, objs...)
	}
}

// ObjectsRemoved implements ObjectsChangeListener.
func (c *ObjectsChangeCaster) ObjectsRemoved(cc fx.ControlContext, objs ...Object) {
	for _, ln := range c.listeners {
		ln.ObjectsRemoved(cc, objs...)
	}
}

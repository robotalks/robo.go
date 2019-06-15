// Package see is the adapter to visualize a 2D world in
// github.com/robotalks/see.
package see

import (
	"encoding/json"
	"fmt"

	fx "github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/sim"
)

// Adapter is the visualization adapter to visualize using
// github.com/robotalks/see.
type Adapter struct {
	Config *Config
	Mapper ObjectMapper

	initial    bool
	updated    map[string]sim.Object
	removedIDs map[string]bool
}

// NewAdapter creates the adapter.
func NewAdapter(config *Config) *Adapter {
	return &Adapter{
		Config:  config,
		initial: true,
	}
}

// Subscribe is a helper to subscribe object changes.
func (a *Adapter) Subscribe(sub sim.ObjectsChangeSubscriber) *Adapter {
	sub.SubscribeObjectsChange(a)
	return a
}

// ObjectsChanged implements ObjectsChangeListener.
func (a *Adapter) ObjectsChanged(cc fx.ControlContext, objs ...sim.Object) {
	if a.updated == nil {
		a.updated = make(map[string]sim.Object)
	}
	for _, obj := range objs {
		a.updated[obj.Name()] = obj
		if a.removedIDs != nil {
			delete(a.removedIDs, obj.Name())
		}
	}
}

// ObjectsRemoved implements ObjectsChangeListener.
func (a *Adapter) ObjectsRemoved(cc fx.ControlContext, objs ...sim.Object) {
	if a.removedIDs == nil {
		a.removedIDs = make(map[string]bool)
	}
	for _, obj := range objs {
		a.removedIDs[obj.Name()] = true
		if a.updated != nil {
			delete(a.updated, obj.Name())
		}
	}
}

// AddToLoop implements LoopAdder.
func (a *Adapter) AddToLoop(l *fx.Loop) {
	l.AddController(fx.PrLvPostProc, fx.ControlFunc(a.ReportChanges))
}

// ReportChanges is a controller to report changes.
func (a *Adapter) ReportChanges(cc fx.ControlContext) error {
	var msgs []Message
	if a.initial {
		msgs = []Message{
			{Action: ActionReset},
			{Action: ActionObject, Object: NewObject("corner", "corner-lt").With("loc", "lt").At(-a.Config.W/2, -a.Config.H/2).Radius(1)},
			{Action: ActionObject, Object: NewObject("corner", "corner-lb").With("loc", "lb").At(-a.Config.W/2, a.Config.H/2).Radius(1)},
			{Action: ActionObject, Object: NewObject("corner", "corner-rt").With("loc", "rt").At(a.Config.W/2, -a.Config.H/2).Radius(1)},
			{Action: ActionObject, Object: NewObject("corner", "corner-rb").With("loc", "rb").At(a.Config.W/2, a.Config.H/2).Radius(1)},
		}
		a.initial = false
		a.removedIDs = nil
	}

	for _, obj := range a.updated {
		if vo, ok := obj.(VisibleObject); ok {
			for _, mapped := range a.Mapper.MapObject(vo) {
				if mapped == nil {
					continue
				}
				msgs = append(msgs, Message{
					Action: ActionObject,
					Object: mapped,
				})
			}
		}
	}

	for id := range a.removedIDs {
		msgs = append(msgs, Message{
			Action:   ActionRemove,
			RemoveID: id,
		})
	}

	a.updated, a.removedIDs = nil, nil
	if len(msgs) > 0 {
		encoded, _ := json.Marshal(msgs)
		fmt.Println(string(encoded) + "\n")
	}
	return nil
}

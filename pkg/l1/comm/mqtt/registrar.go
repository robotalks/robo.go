package mqtt

import (
	"context"
	"encoding/json"

	fx "github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/l1"
	"github.com/robotalks/robo.go/pkg/l1/comm"
)

// Registrar implements l1.Registrar using MQTT.
type Registrar struct {
	Queue *Queue
	Info  l1.ControllerInfo

	metaJSON  string
	registrar comm.Registrar
}

// NewRegistrar creates a Registrar.
func NewRegistrar(brokerURL string, info l1.ControllerInfo) (*Registrar, error) {
	meta, err := json.Marshal(&info.Meta)
	if err != nil {
		panic(err)
	}
	opts, topicPrefix, err := ClientOptionsFromURL(brokerURL)
	if err != nil {
		return nil, err
	}
	opts.SetBinaryWill(topicPrefix+info.Ref.Name()+"/meta", nil, 1, true)
	if opts.ClientID == "" {
		opts.SetClientID("robo:" + info.Ref.Name())
	}
	r := &Registrar{
		Queue:    NewQueue(opts, topicPrefix),
		Info:     info,
		metaJSON: string(meta),
	}
	r.Queue.OnConnect = func(*Queue) { r.onConnected() }
	r.registrar.Init(NewPacketReadWriter(r.Queue).ForController(info.Ref))
	return r, nil
}

// SendEvent implements Registrar.
func (r *Registrar) SendEvent(ctx context.Context, msg fx.Message) error {
	return r.registrar.SendEvent(ctx, msg)
}

// AddToLoop implements LoopAdder.
func (r *Registrar) AddToLoop(loop *fx.Loop) {
	loop.Add(&r.registrar)
	loop.AddRunnable(r)
}

// Run implements Runnable.
func (r *Registrar) Run(ctx context.Context) error {
	r.Queue.Connect()
	<-ctx.Done()
	r.Queue.PubWith(r.Info.Ref.Name()+"/meta", nil, 1, true)
	r.Queue.Close()
	return nil
}

func (r *Registrar) onConnected() {
	r.Queue.PubWith(r.Info.Ref.Name()+"/meta", []byte(r.metaJSON), 1, true)
}

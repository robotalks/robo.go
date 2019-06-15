package joystick

import (
	"context"
	"log"
	"math"
	"time"

	fx "github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/joystick/device"
	"github.com/robotalks/robo.go/pkg/joystick/msgs"
	"github.com/robotalks/robo.go/pkg/l1"
	connenv "github.com/robotalks/robo.go/pkg/l1/env/connector"
	env "github.com/robotalks/robo.go/pkg/l1/env/controller"
	l1msgs "github.com/robotalks/robo.go/pkg/l1/msgs"
)

// Controller is an L2 controller which sends commands to
// an L1 controller.
type Controller struct {
	Env         *env.Env
	DeviceIndex int
	Verbose     bool

	conn        *connection
	eventCh     chan device.Event
	device      device.Device
	deviceTimer <-chan time.Time

	status        msgs.JoystickStatus
	statusChanged bool
}

// NewController creates a Controller.
func NewController(e *env.Env) *Controller {
	return &Controller{
		Env:           e,
		DeviceIndex:   defaultConfig.DeviceIndex,
		Verbose:       defaultConfig.Verbose,
		statusChanged: true,
	}
}

// AddToLoop implements LoopAdder.
func (c *Controller) AddToLoop(loop *fx.Loop) {
	loop.AddRunnable(c)
	loop.AddController(fx.PrLvControl, c)
	loop.AddController(fx.PrLvPostProc, fx.ControlFunc(c.notifyStatusChange))
}

// Run implements Runnable.
func (c *Controller) Run(ctx context.Context) error {
	defer func() {
		if c.device != nil {
			c.device.Close()
		}
	}()
	loopCtl := fx.LoopCtlFrom(ctx)
	c.deviceTimer = time.After(time.Second)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.deviceTimer:
			c.deviceTimer = nil
			var js device.Device
			var err error
			if c.DeviceIndex >= 0 {
				if js, err = device.Open(c.DeviceIndex); err != nil {
					log.Printf("Open joystick %d error: %v", c.DeviceIndex, err)
				}
			} else {
				log.Println("Detecting joystick ...")
				if js, err = device.DetectAndOpen(0); err != nil {
					log.Printf("Detect joystick error: %v", err)
				} else if js == nil {
					log.Printf("No joystick detected.")
				}
			}
			if err == nil && js != nil {
				log.Printf("Joystick %d %q opened!", js.Index(), js.Name())
				c.device, c.eventCh = js, make(chan device.Event, 1)
				go c.pollJoystick(ctx)
				loopCtl.PostMessage(&statusMsg{
					device: &msgs.JoystickDevice{
						Index: uint32(js.Index()),
						Name:  js.Name(),
					},
				})
			} else {
				c.deviceTimer = time.After(time.Second)
			}
		case ev, ok := <-c.eventCh:
			if ok {
				loopCtl.PostMessage(&eventMsg{event: ev})
			} else {
				loopCtl.PostMessage(&eventMsg{stopAll: true})
				if c.device != nil {
					c.device.Close()
				}
				c.device, c.eventCh = nil, nil
				c.deviceTimer = time.After(time.Second)
				loopCtl.PostMessage(&statusMsg{
					device: &msgs.JoystickDevice{Index: 0xffffffff},
				})
			}
			loopCtl.TriggerNext()
		}
	}
}

// Control implements Controller.
func (c *Controller) Control(cc fx.ControlContext) error {
	cc.Messages().ProcessMessages(fx.ProcessMessageFunc(func(mctx fx.MessageProcessingContext) {
		switch msg := mctx.CurrentMessage().(type) {
		case *l1.CommandMsg:
			switch m := msg.Command.Msg().(type) {
			case *msgs.JoystickStatusQuery:
				mctx.MessageTaken()
				msg.Command.Done(&msgs.JoystickStatusReply{Status: &c.status})
			case *msgs.JoystickConnect:
				mctx.MessageTaken()
				msg.Command.Done(c.connect(cc, m))
			}
		case *eventMsg:
			mctx.MessageTaken()
			if conn := c.conn; conn != nil {
				conn.loop.PostMessage(msg)
				conn.loop.TriggerNext()
			} else {
				log.Println("Controller not connected.")
			}
		case *statusMsg:
			if msg.device != nil {
				if msg.device.Index == 0xffffffff {
					c.status.Device = nil
				} else {
					c.status.Device = msg.device
				}
				c.statusChanged = true
			}
			if msg.conn != nil {
				if msg.conn.Type == "" {
					c.status.Connection = nil
				} else {
					c.status.Connection = msg.conn
				}
				c.statusChanged = true
			}
		}
	}))
	return nil
}

func (c *Controller) notifyStatusChange(cc fx.ControlContext) error {
	changed := c.statusChanged
	c.statusChanged = false
	if changed {
		return c.Env.Registrar.SendEvent(cc.Context(), &c.status)
	}
	return nil
}

func (c *Controller) connect(cc fx.ControlContext, msg *msgs.JoystickConnect) fx.Message {
	if c.conn != nil {
		c.conn.close()
		c.conn = nil
		cc.PostMessage(&statusMsg{conn: &msgs.JoystickConnect{}})
	}
	if msg.Type == "" && msg.ID == "" {
		// treat as disconnect.
		return l1msgs.NewCommandOK()
	}
	conf := connenv.NewConfig()
	if conf.RegistryURL = msg.RegistryURL; conf.RegistryURL == "" {
		conf.RegistryURL = c.Env.RegistryURLs[0]
	}
	if conf.Ref.Type, conf.Ref.ID = msg.Type, msg.ID; !conf.Ref.IsValid() {
		return l1msgs.NewCommandErrFromMsg("controller ref invalid")
	}
	connector, err := conf.NewConnector()
	if err != nil {
		return l1msgs.NewCommandErr(err)
	}
	if c.conn, err = newConnection(cc, connector, conf.Ref); err != nil {
		return l1msgs.NewCommandErr(err)
	}
	go c.conn.run()
	cc.PostMessage(&statusMsg{conn: &msgs.JoystickConnect{
		RegistryURL: conf.RegistryURL,
		Type:        conf.Ref.Type,
		ID:          conf.Ref.ID,
	}})
	return l1msgs.NewCommandOK()
}

func (c *Controller) pollJoystick(ctx context.Context) {
	dev, ch := c.device, c.eventCh
	defer close(ch)
	for {
		ev, err := dev.ReadEvent()
		if err != nil {
			log.Printf("Joystick read error: %v", err)
			return
		}
		if ev != nil {
			if c.Verbose {
				var prefix string
				if ev.IsInit() {
					prefix = "[INIT] "
				}
				switch evt := ev.(type) {
				case device.AxisEvent:
					log.Printf(prefix+"Axis %d: %d", evt.Index(), evt.Value())
				case device.ButtonEvent:
					log.Printf(prefix+"Button %d: %v", evt.Index(), evt.Pressed())
				}
			}
			ch <- ev
		}
	}
}

type statusMsg struct {
	device *msgs.JoystickDevice
	conn   *msgs.JoystickConnect
}

func (m *statusMsg) NewMessage() fx.Message { return &statusMsg{} }

type eventMsg struct {
	event   device.Event
	stopAll bool
}

func (m *eventMsg) NewMessage() fx.Message { return &eventMsg{} }

type capsMsg struct {
	caps *l1msgs.Nav2DCaps
}

func (m *capsMsg) NewMessage() fx.Message { return &capsMsg{} }

type connection struct {
	ctx    context.Context
	cancel func()
	conn   l1.ControllerConn
	loop   *fx.Loop
	caps   *l1msgs.Nav2DCaps
}

func newConnection(cc fx.ControlContext, connector l1.Connector, ref l1.ControllerRef) (c *connection, err error) {
	c = &connection{}
	c.ctx, c.cancel = context.WithCancel(cc.Context())
	if c.conn, err = connector.Connect(c.ctx, ref); err != nil {
		return
	}
	c.loop = fx.NewLoop()
	if adder, ok := c.conn.(fx.LoopAdder); ok {
		c.loop.Add(adder)
	}
	c.loop.AddController(fx.PrLvControl, c)
	return
}

func (c *connection) run() {
	c.loop.Run(c.ctx)
}

func (c *connection) close() {
	c.cancel()
}

func (c *connection) handleEvent(ev device.Event) {
	if c.caps == nil {
		log.Println("Nav2DCaps not available.")
		return
	}
	if axisEv, ok := ev.(device.AxisEvent); ok {
		switch axisEv.Index() {
		case 0, 6:
			c.turn(axisEv.Value())
		case 1, 7:
			c.drive(-axisEv.Value())
		}
	}
}

func (c *connection) drive(val int) {
	if c.caps.DriveSpeedMax == 0 {
		log.Println("Nav2DCaps.drive_speed_max not set")
	}
	var msg l1msgs.Nav2DDrive
	msg.Speed = c.caps.DriveSpeedMax * float32(val) / 32767
	c.conn.DoCommand(&msg)
}

func (c *connection) turn(val int) {
	maxVal := c.caps.TurnSpeedMax
	if maxVal == 0 {
		maxVal = math.Pi
	}
	var msg l1msgs.Nav2DTurn
	msg.Speed = maxVal * float32(val) / 32767
	c.conn.DoCommand(&msg)
}

func (c *connection) stopAll() {
	c.conn.DoCommand(&l1msgs.Nav2DDrive{})
	c.conn.DoCommand(&l1msgs.Nav2DTurn{})
}

// Run implements Runnable to query Nav2DCaps from the controller.
func (c *connection) Run(ctx context.Context) error {
	for {
		res := <-c.conn.DoCommand(&l1msgs.Nav2DCapsQuery{}).ResultChan()
		if res.Err != nil {
			log.Printf("Nav2DCapsQuery error: %v", res.Err)
		} else if caps, ok := res.Msg.(*l1msgs.Nav2DCaps); ok {
			loopCtl := fx.LoopCtlFrom(ctx)
			loopCtl.PostMessage(&capsMsg{caps: caps})
			loopCtl.TriggerNext()
			break
		} else {
			log.Println("Nav2DCapsQuery got unknown response")
		}
		<-time.After(time.Second)
	}
	return nil
}

// Control implements Controller.
func (c *connection) Control(cc fx.ControlContext) error {
	cc.Messages().ProcessMessages(fx.ProcessMessageFunc(func(mctx fx.MessageProcessingContext) {
		switch msg := mctx.CurrentMessage().(type) {
		case *eventMsg:
			mctx.MessageTaken()
			if msg.stopAll {
				c.stopAll()
			} else {
				c.handleEvent(msg.event)
			}
		case *capsMsg:
			mctx.MessageTaken()
			log.Printf("Nav2DCaps available: %s", msg.caps.String())
			c.caps = msg.caps
		}
	}))
	return nil
}

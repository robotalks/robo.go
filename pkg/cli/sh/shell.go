package sh

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/abiosoft/ishell"

	fx "github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/l1"
	env "github.com/robotalks/robo.go/pkg/l1/env/connector"
	"github.com/robotalks/robo.go/pkg/l1/msgs"
)

// Shell provides ishell backed interactive shell.
type Shell struct {
	Interactive bool
	OutputJSON  bool
	AutoConnect bool

	Shell  *ishell.Shell
	Config *env.Config
	Loop   *ConnLoop
}

// ConnLoop is a running loop with a controller connection.
type ConnLoop struct {
	Ctx    context.Context
	Cancel func()
	Ref    l1.ControllerRef
	Loop   *fx.Loop
	Conn   l1.ControllerConn
}

const (
	shellKey          = "$shell"
	unconnectedPrompt = "[none] > "
)

var (
	// flags

	evalOnly   bool
	outputJSON bool

	// commands
	commands = []*ishell.Cmd{
		&DiscoverCmd,
		&ConnectCmd,
		&DisconnectCmd,
	}
)

func init() {
	flag.BoolVar(&evalOnly, "e", evalOnly, "Evaluation only, no interactive shell.")
	flag.BoolVar(&outputJSON, "json", outputJSON, "Print output in JSON.")
}

// AddCmds is used by other commands providers during init func.
func AddCmds(cmds ...*ishell.Cmd) {
	commands = append(commands, cmds...)
}

// New creates a new shell.
func New(conf *env.Config) *Shell {
	s := &Shell{
		Interactive: !evalOnly,
		OutputJSON:  outputJSON,

		Shell:  ishell.New(),
		Config: conf,
	}
	s.Shell.Set(shellKey, s)
	s.Shell.SetPrompt(unconnectedPrompt)
	for _, cmd := range commands {
		s.Shell.AddCmd(cmd)
	}
	return s
}

// ShellFrom gets Shell from ishell context.
func ShellFrom(c *ishell.Context) *Shell {
	return c.Get(shellKey).(*Shell)
}

// MustBeConnected wraps command func requires a connection.
func MustBeConnected(fn func(c *ishell.Context)) func(c *ishell.Context) {
	return func(c *ishell.Context) {
		if ShellFrom(c).Loop == nil {
			c.Err(fmt.Errorf("not connected"))
			return
		}
		fn(c)
	}
}

// FormatInfo prints ControllerInfo into friendly string for display.
func FormatInfo(info l1.ControllerInfo) string {
	var w bytes.Buffer
	fmt.Fprintf(&w, "%s", info.Ref.Name())
	if info.Meta.Description != "" {
		fmt.Fprintf(&w, ": %s", info.Meta.Description)
	}
	return w.String()
}

// DoCommand runs a command and waits for result.
func DoCommand(c *ishell.Context, msg fx.Message) (err error) {
	s := ShellFrom(c)
	if s.Loop == nil {
		err = fmt.Errorf("not connected")
		c.Err(err)
		return
	}
	f := s.Loop.Conn.DoCommand(msg)
	select {
	case res := <-f.ResultChan():
		if res.Err != nil {
			c.Err(res.Err)
			return res.Err
		}
		if s.OutputJSON {
			out, err := json.Marshal(res.Msg.(msgs.SerializableMessage).Serializable())
			if err != nil {
				c.Err(err)
				return err
			}
			c.Println(string(out))
			return nil
		}
		if _, ok := res.Msg.(*msgs.CommandOK); ok {
			c.Println("OK")
			return nil
		}
		c.Printf("%s %s\n",
			reflect.Indirect(reflect.ValueOf(res.Msg)).Type().Name(),
			res.Msg.(msgs.SerializableMessage).Serializable().String())
	case <-time.After(time.Second):
		c.Err(fmt.Errorf("Command timeout"))
		return context.DeadlineExceeded
	}
	return nil
}

// WithAutoConnect sets AutoConnect.
func (s *Shell) WithAutoConnect(en bool) *Shell {
	s.AutoConnect = en
	return s
}

// DiscoverControllers discovers controllers.
func (s *Shell) DiscoverControllers(filter func(l1.ControllerInfo) bool) (l1.Connector, []l1.ControllerInfo, error) {
	connector, err := s.Config.NewConnector()
	if err != nil {
		return nil, nil, err
	}
	infoList, err := connector.Discover(context.TODO())
	if err != nil {
		return connector, nil, err
	}
	if filter != nil {
		items := make([]l1.ControllerInfo, 0, len(infoList))
		for _, info := range infoList {
			if filter(info) {
				items = append(items, info)
			}
		}
		infoList = items
	}
	return connector, infoList, nil
}

// SelectController discovers controllers and asks for a choice.
func (s *Shell) SelectController(filter func(l1.ControllerInfo) bool) (l1.Connector, *l1.ControllerInfo, error) {
	connector, infoList, err := s.DiscoverControllers(filter)
	if err != nil {
		return nil, nil, err
	}
	if len(infoList) == 0 {
		return connector, nil, nil
	}
	var index int
	if len(infoList) > 1 {
		if !s.Interactive {
			return nil, nil, fmt.Errorf("more than 1 controllers discovered in non-interactive mode")
		}
		items := make([]string, len(infoList))
		for n, info := range infoList {
			items[n] = info.Ref.Name()
			if info.Meta.Description != "" {
				items[n] += ": " + info.Meta.Description
			}
		}
		index = s.Shell.MultiChoice(items, "Which one to connect?")
	}

	return connector, &infoList[index], nil
}

// Connect connects controller with ref.
func (s *Shell) Connect(ref l1.ControllerRef) error {
	connector, err := s.Config.NewConnector()
	if err != nil {
		return err
	}
	connLoop := &ConnLoop{Ref: ref}
	connLoop.Ctx, connLoop.Cancel = context.WithCancel(context.Background())
	if connLoop.Conn, err = connector.Connect(connLoop.Ctx, ref); err != nil {
		return err
	}
	connLoop.Loop = fx.NewLoop()
	if adder, ok := connLoop.Conn.(fx.LoopAdder); ok {
		connLoop.Loop.Add(adder)
	}
	if s.Loop != nil {
		s.Loop.Cancel()
	}
	s.Loop = connLoop
	go connLoop.Loop.Run(connLoop.Ctx)
	s.Shell.SetPrompt(fmt.Sprintf("%s > ", ref.Name()))
	return nil
}

// Disconnect disconnects current controller.
func (s *Shell) Disconnect() {
	if s.Loop != nil {
		s.Loop.Cancel()
		s.Loop = nil
		s.Shell.SetPrompt(unconnectedPrompt)
	}
}

// Run runs the shell.
func (s *Shell) Run(args ...string) {
	if s.AutoConnect && s.Config.Ref.IsValid() {
		if s.Interactive {
			s.Shell.Printf("Connecting %s ...\n", s.Config.Ref.Name())
		}
		if err := s.Connect(s.Config.Ref); err != nil {
			log.Fatalf("connect %q failed: %v", s.Config.Ref.Name(), err)
		}
	}

	if len(args) > 0 {
		if err := s.Shell.Process(args...); err != nil {
			log.Fatalln(err)
		}
		return
	}
	if s.Interactive {
		s.Shell.Run()
		return
	}
	log.Fatalln("command expected")
}

var (
	// DiscoverCmd discovers controllers.
	DiscoverCmd = ishell.Cmd{
		Name:    "discover",
		Aliases: []string{"list", "l"},
		Help:    "",
		Func: func(c *ishell.Context) {
			s := ShellFrom(c)
			_, infoList, err := s.DiscoverControllers(nil)
			if err != nil {
				c.Err(err)
				return
			}
			if s.OutputJSON {
				if len(infoList) == 0 {
					// in case infoList is nil, make it empty slice.
					infoList = []l1.ControllerInfo{}
				}
				out, err := json.Marshal(infoList)
				if err != nil {
					c.Err(err)
					return
				}
				c.Println(string(out))
				return
			}
			if len(infoList) == 0 {
				c.Println("No controllers found")
				return
			}
			for _, info := range infoList {
				c.Println(FormatInfo(info))
			}
		},
	}

	// ConnectCmd connects a controller.
	ConnectCmd = ishell.Cmd{
		Name:    "connect",
		Aliases: []string{"c"},
		Help:    "TYPE ID",
		Func: func(c *ishell.Context) {
			s := ShellFrom(c)
			var ref l1.ControllerRef
			if len(c.Args) >= 2 {
				ref.Type, ref.ID = c.Args[0], c.Args[1]
			} else {
				var filter func(l1.ControllerInfo) bool
				if len(c.Args) == 1 {
					filter = func(info l1.ControllerInfo) bool {
						return info.Ref.Type == c.Args[0]
					}
				}
				_, info, err := s.SelectController(filter)
				if err != nil {
					c.Err(err)
					return
				}
				if info == nil {
					c.Err(fmt.Errorf("no controller discovered"))
					return
				}
				ref = info.Ref
			}
			if err := s.Connect(ref); err != nil {
				c.Err(err)
				return
			}
		},
	}

	// DisconnectCmd disconnects current controller.
	DisconnectCmd = ishell.Cmd{
		Name:    "disconnect",
		Aliases: []string{"d"},
		Help:    "",
		Func: func(c *ishell.Context) {
			ShellFrom(c).Disconnect()
		},
	}
)

// Main is a helper to provide a single call in main.
func Main() {
	flag.Parse()
	New(env.NewConfig()).WithAutoConnect(true).Run(flag.Args()...)
}

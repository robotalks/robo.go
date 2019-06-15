package joystick

import (
	"fmt"

	"github.com/abiosoft/ishell"

	"github.com/robotalks/robo.go/pkg/cli/sh"
	"github.com/robotalks/robo.go/pkg/joystick/msgs"
	"github.com/robotalks/robo.go/pkg/l1"
)

var (
	// JoystickStatusCmd exposes JoystickStatusQuery command.
	JoystickStatusCmd = ishell.Cmd{
		Name:    "js.status",
		Aliases: []string{"jss"},
		Help:    "",
		Func: sh.MustBeConnected(func(c *ishell.Context) {
			sh.DoCommand(c, &msgs.JoystickStatusQuery{})
		}),
	}

	// JoystickConnectCmd exposes JoystickConnect command.
	JoystickConnectCmd = ishell.Cmd{
		Name:    "js.connect",
		Aliases: []string{"jsc"},
		Help:    "TYPE ID [REGISTRY_URL]",
		Func: sh.MustBeConnected(func(c *ishell.Context) {
			var msg msgs.JoystickConnect
			if len(c.Args) >= 2 {
				msg.Type, msg.ID = c.Args[0], c.Args[1]
				if len(c.Args) > 2 {
					msg.RegistryURL = c.Args[2]
				}
			} else {
				var filter func(l1.ControllerInfo) bool
				if len(c.Args) == 1 {
					filter = func(info l1.ControllerInfo) bool {
						return info.Ref.Type == c.Args[0]
					}
				}
				_, info, err := sh.ShellFrom(c).SelectController(filter)
				if err != nil {
					c.Err(err)
					return
				}
				if info == nil {
					c.Err(fmt.Errorf("no controller discovered"))
					return
				}
				msg.Type, msg.ID = info.Ref.Type, info.Ref.ID
			}
			sh.DoCommand(c, &msg)
		}),
	}
)

func init() {
	sh.AddCmds(
		&JoystickStatusCmd,
		&JoystickConnectCmd,
	)
}

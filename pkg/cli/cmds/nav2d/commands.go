package nav2d

import (
	"fmt"
	"strconv"

	"github.com/abiosoft/ishell"

	"github.com/robotalks/robo.go/pkg/cli/sh"
	"github.com/robotalks/robo.go/pkg/l1/msgs"
	"github.com/robotalks/robo.go/pkg/sim"
)

var (
	// Nav2DCapsQueryCmd exposes Nav2DCapsQuery command.
	Nav2DCapsQueryCmd = ishell.Cmd{
		Name:    "nav2d.caps",
		Aliases: []string{"n2caps"},
		Help:    "",
		Func: sh.MustBeConnected(func(c *ishell.Context) {
			sh.DoCommand(c, &msgs.Nav2DCapsQuery{})
		}),
	}

	// Nav2DDriveCmd exposes Nav2DDrive command.
	Nav2DDriveCmd = ishell.Cmd{
		Name:    "nav2d.drive",
		Aliases: []string{"n2d"},
		Help:    "SPEED(mm/s) [ACCEL(mm/s^2)]",
		Func: sh.MustBeConnected(func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(fmt.Errorf("SPEED required"))
				return
			}
			var msg msgs.Nav2DDrive
			val, err := strconv.ParseFloat(c.Args[0], 32)
			if err != nil {
				c.Err(fmt.Errorf("Invalid SPEED: %v", err))
				return
			}
			msg.Speed = float32(val)
			if len(c.Args) > 1 {
				val, err = strconv.ParseFloat(c.Args[1], 32)
				if err != nil {
					c.Err(fmt.Errorf("Invalid ACCEL: %v", err))
					return
				}
				msg.Accelation = float32(val)
			}
			sh.DoCommand(c, &msg)
		}),
	}

	// Nav2DTurnCmd exposes Nav2DTurn command.
	Nav2DTurnCmd = ishell.Cmd{
		Name:    "nav2d.turn",
		Aliases: []string{"n2t"},
		Help:    "SPEED(degrees/s)",
		Func: sh.MustBeConnected(func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Err(fmt.Errorf("SPEED required"))
				return
			}
			var msg msgs.Nav2DTurn
			val, err := strconv.ParseFloat(c.Args[0], 32)
			if err != nil {
				c.Err(fmt.Errorf("Invalid SPEED: %v", err))
				return
			}
			msg.Speed = float32(sim.AngleFromDegrees(val).Radians())
			sh.DoCommand(c, &msg)
		}),
	}
)

func init() {
	sh.AddCmds(
		&Nav2DCapsQueryCmd,
		&Nav2DDriveCmd,
		&Nav2DTurnCmd,
	)
}

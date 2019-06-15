package nav

import (
	"flag"

	env "github.com/robotalks/robo.go/pkg/l1/env/controller"
	"github.com/robotalks/robo.go/pkg/sim"
)

// Config defines the configuration for the bot.
type Config struct {
	Size          float64
	DriveSpeedMax float64
	TurnSpeedMax  float64
}

// Defaults
const (
	DefaultSize          float64 = 50
	DefaultDriveSpeedMax float64 = 500
)

var defaultConfig = Config{
	Size:          DefaultSize,
	DriveSpeedMax: DefaultDriveSpeedMax,
}

// SetupFlags sets command line flags.
func SetupFlags() {
	flag.Float64Var(&defaultConfig.Size, "bot-size", defaultConfig.Size, "Size (mm) of the bot, it's square.")
	flag.Float64Var(&defaultConfig.DriveSpeedMax, "drive-speed-max", defaultConfig.DriveSpeedMax, "Maximum drive speed (mm/s).")
	flag.Float64Var(&defaultConfig.TurnSpeedMax, "turn-speed-max", defaultConfig.TurnSpeedMax, "Maximum turn speed (degrees/s), 0 means unlimited.")
}

// Default gets default config.
func Default() *Config {
	return &defaultConfig
}

// NewConfig creates the default configuration.
func NewConfig() *Config {
	conf := defaultConfig
	return &conf
}

// NewController creates the Controller.
func (c *Config) NewController(e *env.Env) *Controller {
	ctl := NewController(e)
	ctl.Outline.CX, ctl.Outline.CY = c.Size, c.Size
	ctl.Outline.X, ctl.Outline.Y = -ctl.Outline.CX/2, -ctl.Outline.CY/2
	ctl.Nav.Caps.DriveSpeedMax = float32(c.DriveSpeedMax)
	if rad := sim.AngleFromDegrees(c.TurnSpeedMax).Radians(); rad != 0 {
		ctl.Nav.Caps.TurnSpeedMax = float32(rad)
	}
	return ctl
}

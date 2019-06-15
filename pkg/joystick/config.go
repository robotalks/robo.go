package joystick

import (
	"flag"

	env "github.com/robotalks/robo.go/pkg/l1/env/controller"
)

// Config defines the configurations for the controller.
type Config struct {
	DeviceIndex int
	Verbose     bool
}

var defaultConfig = Config{
	DeviceIndex: -1,
}

// SetupFlags sets command line flags.
func SetupFlags() {
	flag.IntVar(&defaultConfig.DeviceIndex, "device", defaultConfig.DeviceIndex, "Device index, -1 for auto detection.")
	flag.BoolVar(&defaultConfig.Verbose, "verbose", defaultConfig.Verbose, "Print Joystick events.")
}

// Default gets default config.
func Default() *Config {
	return &defaultConfig
}

// NewConfig creates a config with defaults.
func NewConfig() *Config {
	conf := defaultConfig
	return &conf
}

// NewController creates a controller using the config.
func (c *Config) NewController(e *env.Env) *Controller {
	ctl := NewController(e)
	ctl.DeviceIndex = c.DeviceIndex
	ctl.Verbose = c.Verbose
	return ctl
}

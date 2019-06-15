package see

import "flag"

// Config represents configuration for see.
type Config struct {
	W float64
	H float64
}

var defaultConfig = Config{
	W: 1000,
	H: 1000,
}

// SetupFlags sets command line flags.
func SetupFlags() {
	flag.Float64Var(&defaultConfig.W, "see-w", defaultConfig.W, "Width (mm) of visualization area")
	flag.Float64Var(&defaultConfig.H, "see-h", defaultConfig.H, "Height (mm) of visualization area")
}

// Default gets default config.
func Default() *Config {
	return &defaultConfig
}

// NewConfig creates a default config.
func NewConfig() *Config {
	conf := defaultConfig
	return &conf
}

// NewAdapter creates adapter from config.
func (c *Config) NewAdapter() *Adapter {
	return NewAdapter(c)
}

package controller

import (
	"flag"
	"fmt"
	"log"
	"os"

	fx "github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/l1"
	"github.com/robotalks/robo.go/pkg/l1/comm"
	"github.com/robotalks/robo.go/pkg/l1/comm/mqtt"
	"github.com/robotalks/robo.go/pkg/l1/env"
)

// Config provides common options to setup an env for L1 controllers.
type Config struct {
	Info l1.ControllerInfo

	// MQTTBrokerURL specifies the MQTT broker to use.
	// e.g. mqtt://host:port/topic-prefix
	MQTTBrokerURL string
}

var defaultConfig = Config{
	MQTTBrokerURL: "mqtt://localhost:1883/robo/",
}

func init() {
	if val := os.Getenv("ROBO_MQTT_URL"); val != "" {
		defaultConfig.MQTTBrokerURL = val
	}
	defaultConfig.Info.Ref.ID = env.MachineID()
}

// SetupFlags sets command line flags.
func SetupFlags() {
	flag.StringVar(&defaultConfig.Info.Ref.Type, "type", defaultConfig.Info.Ref.Type, "Controller type")
	flag.StringVar(&defaultConfig.Info.Ref.ID, "id", defaultConfig.Info.Ref.ID, "Controller ID")
	flag.StringVar(&defaultConfig.MQTTBrokerURL, "mqtt", defaultConfig.MQTTBrokerURL, "MQTT broker URL")
}

// Default gets default config.
func Default() *Config {
	return &defaultConfig
}

// SetControllerType should be called in init with basic info about the controller.
func SetControllerType(typ string, meta l1.ControllerMeta) {
	defaultConfig.Info.Ref.Type = typ
	defaultConfig.Info.Meta = meta
}

// Env is the env for L1 controllers.
type Env struct {
	Config       *Config
	RegistryURLs []string
	Registrar    *comm.RegistrarMux
}

// NewConfig creates a Config with default configurations.
func NewConfig() *Config {
	conf := defaultConfig
	return &conf
}

// NewEnv creates Env from config.
func (c *Config) NewEnv() (*Env, error) {
	if !c.Info.Ref.IsValid() {
		return nil, fmt.Errorf("robot type and id must be specified")
	}
	env := &Env{
		Config:    c,
		Registrar: &comm.RegistrarMux{},
	}
	if c.MQTTBrokerURL != "" {
		reg, err := mqtt.NewRegistrar(c.MQTTBrokerURL, c.Info)
		if err != nil {
			return nil, fmt.Errorf("create MQTT registrar error: %v", err)
		}
		env.Registrar.Add(reg)
		env.RegistryURLs = append(env.RegistryURLs, c.MQTTBrokerURL)
	}
	if len(env.Registrar.Registrars) == 0 {
		return nil, fmt.Errorf("at least one registrar is required")
	}
	return env, nil
}

// MustNewEnv creates Env and fails on error.
func (c *Config) MustNewEnv() *Env {
	env, err := c.NewEnv()
	if err != nil {
		log.Fatalln(err)
	}
	return env
}

// AddToLoop adds controllers/runners to loop.
func (e *Env) AddToLoop(loop *fx.Loop) {
	loop.Add(e.Registrar)
	loop.Add(&comm.UnsupportedCommands{})
}

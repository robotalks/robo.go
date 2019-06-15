package connector

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/robotalks/robo.go/pkg/l1"
	"github.com/robotalks/robo.go/pkg/l1/comm/mqtt"
)

// Config provides common options to setup Connectors.
type Config struct {
	Ref l1.ControllerRef

	// RegistryURL specifies the URL of controller registry.
	// e.g. mqtt://host:port/topic-prefix
	RegistryURL string
}

var defaultConfig = Config{
	RegistryURL: "mqtt://localhost:1883/robo/",
}

func init() {
	if val := os.Getenv("ROBO_TYPE"); val != "" {
		defaultConfig.Ref.Type = val
	}
	if val := os.Getenv("ROBO_ID"); val != "" {
		defaultConfig.Ref.ID = val
	}
	if val := os.Getenv("ROBO_REGISTRY_URL"); val != "" {
		defaultConfig.RegistryURL = val
	}
}

// SetupFlags sets up command line flags.
func SetupFlags() {
	flag.StringVar(&defaultConfig.Ref.Type, "robot-type", defaultConfig.Ref.Type, "Robot type to connect.")
	flag.StringVar(&defaultConfig.Ref.ID, "robot-id", defaultConfig.Ref.ID, "Robot ID to connect.")
	flag.StringVar(&defaultConfig.RegistryURL, "robot-reg", defaultConfig.RegistryURL, "Robot Registry URL.")
}

// Default gets the default config.
func Default() *Config {
	return &defaultConfig
}

// NewConfig creates a Config with default configurations.
func NewConfig() *Config {
	conf := defaultConfig
	return &conf
}

// NewConnector creates a Connector using current config.
func (c *Config) NewConnector() (l1.Connector, error) {
	parsedURL, err := url.Parse(c.RegistryURL)
	if err != nil {
		return nil, fmt.Errorf("invalid registry URL: %v", err)
	}
	switch parsedURL.Scheme {
	case "mqtt":
		return mqtt.NewConnector(c.RegistryURL)
	default:
		return nil, fmt.Errorf("unknown registry URL scheme: %q", parsedURL.Scheme)
	}
}

// MustNewConnector creates a Connector and fails on error.
func (c *Config) MustNewConnector() l1.Connector {
	conn, err := c.NewConnector()
	if err != nil {
		log.Fatalln(err)
	}
	return conn
}

// Connect directly connects to L1 controller.
func (c *Config) Connect() (l1.ControllerConn, error) {
	if !c.Ref.IsValid() {
		return nil, fmt.Errorf("robot type and id must be specified")
	}
	connector, err := c.NewConnector()
	if err != nil {
		return nil, err
	}
	return connector.Connect(context.TODO(), c.Ref)
}

// MustConnect connects to L1 controller for fail.
func (c *Config) MustConnect() l1.ControllerConn {
	conn, err := c.Connect()
	if err != nil {
		log.Fatalln(err)
	}
	return conn
}

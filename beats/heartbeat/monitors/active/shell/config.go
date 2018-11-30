package shell

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/elastic/beats/heartbeat/monitors"
	"github.com/elastic/beats/libbeat/common/match"
	"github.com/elastic/beats/libbeat/common/transport/tlscommon"
)

type Config struct {
	Name string `config:"name"`

	// connection settings
	Hosts []string `config:"hosts" validate:"required"`

	Mode monitors.IPSettings `config:",inline"`
	// authentication
	Username string `config:"username"`
	Password string `config:"password"`
	Key      string `config:"key"`
	// configure tls
	TLS *tlscommon.Config `config:"ssl"`
	// configure validation
	Check         checkConfig   `config:"check"`
	CustomeFields []string      `config:"custom"`
	Timeout       time.Duration `config:"timeout"`

	Docker       bool     `config:"docker"`
	Dockerfilter []string `config:"dockerfilter"`
}

type checkConfig struct {
	Request  commandConfig `config:"request"`
	Response outputConfig  `config:"output"`
}

type commandConfig struct {
	Command string   `config:"command"`
	Args    []string `config:"args"`
	Dir     string   `config:"dir"`
}

type outputConfig struct {
	Ok       []match.Matcher `config:"ok"`
	Critical []match.Matcher `config:"critical"`
}

// defaultConfig creates a new copy of the monitors default configuration.
func defaultConfig() Config {
	return Config{
		Name:         "echo",
		Hosts:        []string{"localhost:22"},
		Mode:         monitors.DefaultIPSettings,
		TLS:          nil,
		Timeout:      16 * time.Second,
		Docker:       false,
		Dockerfilter: []string{},
		Check: checkConfig{
			Request: commandConfig{
				Dir: "",
			},
			Response: outputConfig{},
		},
	}
}

func (c *Config) Validate() error {

	if c.Docker {
		if len(c.Dockerfilter) == 0 {
			return fmt.Errorf("The dockerfiler is required for docker shell command")
		}
	} else {
		for _, addr := range c.Hosts {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				return err
			}

			if strings.ToLower(host) != "localhost" {
				if c.Username == "" {
					return fmt.Errorf("Username is required")
				}

				if c.Password == "" && c.Key == "" {
					return fmt.Errorf("Either Password and key is required")
				}
				if strings.Index(c.Key, "@") == 0 {
					_, err := os.Stat(string(c.Key[1:]))
					if err != nil {
						return err
					}

				}
			}

		}

	}

	return nil
}

func (c *checkConfig) Validate() error {
	return nil
}

func (c *commandConfig) Validate() error {
	return nil
}

func (c *outputConfig) Validate() error {
	return nil

}

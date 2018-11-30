package shell

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/elastic/beats/heartbeat/monitors"
)

const (
	monitorName = "shell"
	plainScheme = "shell"
)

type Client interface {
	Connect() error
	Reconnect() error
	Close()
	Run(dir, command string, args ...string) (string, error)
}

var debugf = logp.MakeDebug(monitorName)

func init() {
	monitors.RegisterActive(monitorName, create)
}

func create(
	name string,
	cfg *common.Config,
) (jobs []monitors.Job, endpointNum int, err error) {

	// unpack the monitors configuration
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, 0, err
	}

	validator := makeValidator(&config)

	jobs = make([]monitors.Job, len(config.Hosts))

	for i, host := range config.Hosts {
		jobs[i], err = newShellMonitorJob(host, &config, validator)
		if err != nil {
			return nil, 0, err
		}
	}

	return jobs, len(config.Hosts), nil

}

package shell

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/elastic/beats/heartbeat/monitors/active/shell/docker"
	"github.com/elastic/beats/heartbeat/monitors/active/shell/local"
	"github.com/elastic/beats/heartbeat/monitors/active/shell/ssh"

	"github.com/elastic/beats/heartbeat/monitors"
	"github.com/elastic/beats/heartbeat/reason"
	"github.com/elastic/beats/libbeat/common"
)

type shellComm interface {
	Run(dir, command string, args ...string) (string, error)
}

func createClinet(addr string, config *Config) (Client, error) {
	// fmt.Println(addr)
	if config.Docker {
		docker := docker.NewDockerClient()
		docker.Endpoint = addr
		docker.Timeout = config.Timeout
		docker.Filter = config.Dockerfilter
		return docker, nil
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(host) == "localhost" {
		lclient := local.NewLocalClient()
		lclient.Timeout = config.Timeout
		return lclient, nil
	}
	sshClient := ssh.NewSSHClient()
	sshClient.Addr = addr
	sshClient.Username = config.Username
	sshClient.Password = config.Password
	sshClient.Timeout = config.Timeout
	sshClient.Key = config.Key
	return sshClient, nil
}

func newShellMonitorJob(
	addr string,
	config *Config,
	validator OutputCheck,
) (monitors.Job, error) {

	typ := config.Name
	jobName := fmt.Sprintf("%v@%v", typ, addr)

	cmd, err := createClinet(addr, config)
	if err != nil {
		return nil, err
	}
	okstr := ""
	for _, ok := range config.Check.Response.Ok {
		okstr = okstr + ok.String() + ","
	}

	criticalStr := ""
	for _, critical := range config.Check.Response.Critical {
		criticalStr = criticalStr + critical.String() + " ,"
	}

	eventFields := common.MapStr{
		"monitor": common.MapStr{
			"scheme":   plainScheme,
			"command":  config.Check.Request.Command,
			"args":     strings.Join(config.Check.Request.Args, " "),
			"dir":      config.Check.Request.Dir,
			"username": config.Username,
			"docker":   config.Docker,
		},
		"check": common.MapStr{
			"ok":       okstr,
			"critical": criticalStr,
		},
	}

	customs := config.CustomeFields
	if len(customs) != 0 {
		customMap := common.MapStr{}
		for _, v := range customs {
			splitPos := strings.Index(v, ":")
			if splitPos > 0 && splitPos != len(v)-1 {
				customMap[string(v[0:splitPos])] = string(v[splitPos+1:])
			}
		}
		if len(customMap) > 0 {
			eventFields["custom"] = customMap
		}
	}

	settings := monitors.MakeJobSetting(jobName).WithFields(eventFields)

	return monitors.MakeSimpleJob(settings, func() (common.MapStr, error) {

		_, _, event, err := runCommand(cmd, config.Check.Request.Dir, config.Check.Request.Command, validator, config.Check.Request.Args...)
		return event, err
	}), nil
}

func runCommand(comm shellComm, dir, command string, validate func(string) error, args ...string) (start, end time.Time, event common.MapStr, errReason reason.Reason) {
	start = time.Now()
	output, err := comm.Run(dir, command, args...)
	end = time.Now()
	event = makeEvent(output)
	if err == nil {
		err = validate(output)
	}
	errReason = reason.ValidateFailed(err)
	return
}

func makeEvent(output string) common.MapStr {
	return common.MapStr{"shell": common.MapStr{
		"response": common.MapStr{
			"output": output,
		},
	}}
}

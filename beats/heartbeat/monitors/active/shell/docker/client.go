package docker

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	dclient "github.com/docker/docker/client"
)

const (
	execLinuxCommand = "/bin/sh"
)

type DockerClient struct {
	dockerClient     *dclient.Client
	hijackedResponse *types.HijackedResponse
	dockerClientOnce *sync.Once
	dockerClientErr  error

	execClientOnce *sync.Once
	execErr        error
	commandMutex   *sync.RWMutex

	actionMutex *sync.RWMutex

	name     string
	Endpoint string
	Filter   []string
	Timeout  time.Duration
}

func NewDockerClient() *DockerClient {
	d := &DockerClient{}
	d.dockerClientOnce = &sync.Once{}
	d.execClientOnce = &sync.Once{}
	d.actionMutex = &sync.RWMutex{}
	d.commandMutex = &sync.RWMutex{}
	return d
}

func (d *DockerClient) client() error {
	d.dockerClientOnce.Do(func() {
		var err error

		d.dockerClient, err = dclient.NewClient(d.Endpoint, "", nil, nil)
		if err != nil {
			d.dockerClientErr = err
		}

	})
	return d.dockerClientErr
}

//CheckClient check whether the client is connected.
func (d *DockerClient) CheckClient() error {
	d.dockerClientOnce = &sync.Once{}
	return d.client()
}

func (d *DockerClient) Connect() error {
	d.execClientOnce.Do(func() {
		err := d.CheckClient()
		if err != nil {
			d.execErr = err
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), d.Timeout)
		defer cancel()
		client := d.dockerClient
		containerDetail, err := d.getContainerBriefDetails(true)
		d.name = strings.Join(containerDetail.Names, " ")
		if err != nil {
			d.execErr = err
			return
		}
		d.actionMutex.Lock()
		defer d.actionMutex.Unlock()

		resp, err := client.ContainerExecCreate(ctx, containerDetail.ID, types.ExecConfig{AttachStderr: true, AttachStdin: true, AttachStdout: true, Tty: false, Detach: false, Cmd: []string{execLinuxCommand}})
		if err != nil {
			d.execErr = err
			return
		}
		// It might need to be changed to type.ExecConfig for the older docker api
		//https://github.com/docker/docker-ce/commit/016304f34671d1e4b9dbe93130700f6c9d2268f6#diff-fb5277a57b9d79722513466ad8a988da
		attachOutput, err := client.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{Detach: false, Tty: false})
		if err != nil {
			d.execErr = err
			return
		}

		d.hijackedResponse = &attachOutput
		d.execErr = nil

	})

	return d.execErr
}

func (d *DockerClient) Reconnect() error {
	d.execClientOnce = &sync.Once{}
	return d.Connect()
}

func (d *DockerClient) Close() {
	d.execClientOnce = &sync.Once{}
	d.execErr = nil
	if d.hijackedResponse == nil {
		return
	}
	if d.hijackedResponse.Conn == nil || d.hijackedResponse.Reader == nil {
		return
	}
	d.hijackedResponse.CloseWrite()
	d.hijackedResponse.Close()
}

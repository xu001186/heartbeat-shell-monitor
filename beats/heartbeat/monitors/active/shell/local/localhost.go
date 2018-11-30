package local

import (
	"context"
	"os/exec"
	"time"
)

type LocalClient struct {
	Timeout time.Duration
}

func NewLocalClient() *LocalClient {
	return &LocalClient{}
}

func (c *LocalClient) Run(dir, command string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.Output()
	return string(output), err
}

func (c *LocalClient) Connect() error {
	return nil
}
func (c *LocalClient) Reconnect() error {
	return nil
}
func (c *LocalClient) Close() {}

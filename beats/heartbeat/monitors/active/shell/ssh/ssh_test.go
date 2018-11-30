package ssh

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	keydata = `-----BEGIN RSA PRIVATE KEY-----

-----END RSA PRIVATE KEY-----`
)

func Test_run_output(t *testing.T) {
	comm := &SSHClient{
		Addr:       "10.4.92.83:22",
		Username:   "root",
		Password:   "XXX",
		Timeout:    2 * time.Second,
		initClient: &sync.Once{},
	}

	defer comm.Run("~", "rm test.sh")
	comm.Run("~", "echo '#!/bin/sh' >test.sh")
	comm.Run("~", "echo 'echo test test1' >test.sh")
	comm.Run("~", "chmod 755 test.sh")
	out, err := comm.Run("~", "/bin/bash", "test.sh")
	assert.Equal(t, "test test1\n", out)
	assert.NoError(t, err, "Failed by running comm.Output")
}

func Test_timeout(t *testing.T) {
	comm := &SSHClient{
		Addr:       "10.4.92.83:22",
		Username:   "root",
		Password:   "XXX",
		Timeout:    2 * time.Second,
		initClient: &sync.Once{},
	}

	_, err := comm.Run("", "sleep 5")
	assert.EqualError(t, err, "Connection is disconnected by the timeout or lost")
	out, err := comm.Run("", "echo", "test", "test1")
	assert.Equal(t, "test test1", out)
	assert.NoError(t, err, "Failed by running timeout test")
}

func Test_Key(t *testing.T) {
	comm := &SSHClient{
		Addr:     "10.4.92.83:22",
		Username: "root",
		Password: "XXX",
		Key:      keydata,

		initClient: &sync.Once{},
	}

	out, err := comm.Run("", "echo", "test", "test1")
	assert.Equal(t, "test test1", out)
	assert.NoError(t, err, "Failed by running timeout test")
}

func Test_Key_file(t *testing.T) {
	comm := &SSHClient{
		Addr:       "10.4.92.83:22",
		Username:   "root",
		Key:        "@/Users/user/.ssh/id_rsa",
		Timeout:    1 * time.Second,
		initClient: &sync.Once{},
	}

	out, err := comm.Run("", "ps aux | grep ipa | wc -l")
	assert.Equal(t, "test test1", out)
	assert.NoError(t, err, "Failed by running timeout test")
}

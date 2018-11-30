package local

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_run_local_output(t *testing.T) {
	testdata := `#!/bin/sh 
	echo test test1
`
	comm := &LocalClient{
		Timeout: 2 * time.Second,
	}
	folder := "/tmp"
	f, err := os.Create(filepath.Join(folder, "test.sh"))
	defer os.Remove(filepath.Join(folder, "test.sh"))
	assert.NoError(t, err, "Failed to create file")
	f.Write([]byte(testdata))
	f.Close()

	out, err := comm.Run(folder, "/bin/bash", "test.sh")
	assert.Equal(t, "test test1\n", out)
	assert.NoError(t, err, "Failed by running comm.Output")

}

func Test_local_timeout(t *testing.T) {
	testdata := `#!/bin/sh
	echo test test1
	sleep 10
`

	comm := &LocalClient{

		Timeout: 1 * time.Second,
	}

	folder := "/tmp"
	f, err := os.Create(filepath.Join(folder, "test.sh"))
	defer os.Remove(filepath.Join(folder, "test.sh"))
	assert.NoError(t, err, "Failed to create file")
	f.Write([]byte(testdata))
	f.Close()

	_, err = comm.Run(folder, "/bin/bash", "test.sh")
	assert.EqualError(t, err, "signal: killed")
}

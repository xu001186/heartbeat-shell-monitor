package docker

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"time"

	"github.com/elastic/beats/heartbeat/monitors/active/shell/util"
)

/**
	Known Issue:
		If the command containes " && " , the hijack sometimes returns the first part command result ,
		and rest part of commands result return to the next hijack .
**/
func (d *DockerClient) readerToString(reader *bufio.Reader) (string, error) {
	header := make([]byte, 8) // [8]byte{STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4}[]byte{OUTPUT}
	_, err := reader.Read(header)
	if err != nil {

		return "", err
	}
	retType := int(header[0]) // 1 --stdout , 2 -- stderr
	length := binary.BigEndian.Uint32(header[4:])

	output := make([]byte, length)
	if err != nil {
		return "", err
	}

	_, err = reader.Read(output)
	if err != nil {
		return "", err
	}

	if retType == 2 {
		return "", errors.New(string(output))
	} else {
		return string(output), nil
	}
}

//fixArgs , always add a echo '' to avoid the timeout issue , and make sure the args end with '\n'
// https://github.com/moby/moby/issues/37182
func fixArgs(command string, args ...string) (string, []string) {
	command = "echo ` " + command
	if len(args) == 0 {
		args = append(args, " ` \n")
	} else {
		lastItem := args[len(args)-1]
		m, _ := regexp.MatchString(".*\n *$", lastItem)
		if m {
			args = append(args[:len(args)-1])
			args = append(args, lastItem[0:strings.LastIndex(lastItem, "\n")])

		}
		args = append(args, " ` \n")
	}
	return command, args

}


func (d *DockerClient) Run(dir, command string, args ...string) (string, error) {
	d.commandMutex.Lock()
	defer d.commandMutex.Unlock()
	command, args = fixArgs(command, args...)

	err := d.Connect()
	if err != nil {
		fmt.Println("reconnect")
		err = d.Reconnect() // always Reconnect if it's failed in first connect
		if err != nil {
			return "", err
		}
	}
	hijacked := d.hijackedResponse
	if hijacked == nil {
		err = fmt.Errorf("connection is closed  for %v ", d.name)
		d.execErr = err
		return "", err
	}
	if hijacked.Conn == nil || hijacked.Reader == nil {
		err = fmt.Errorf("connection is closed  for %v ", d.name)
		d.execErr = err
		return "", err
	}

	// fmt.Println(util.BuildCmd(dir, command, args...))

	_, err = hijacked.Conn.Write([]byte(util.BuildCmd(dir, command, args...)))
	if err != nil {
		d.execErr = err
		return "", err
	}
	hijacked.Conn.SetDeadline(time.Now().Add(d.Timeout))

	output, err := d.readerToString(hijacked.Reader)
	d.execErr = err
	return strings.Trim(string(output), "\n"), err
}

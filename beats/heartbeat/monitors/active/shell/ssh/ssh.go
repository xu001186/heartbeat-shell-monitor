package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/elastic/beats/heartbeat/monitors/active/shell/util"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	sshclient  *ssh.Client
	sshError   error
	initClient *sync.Once

	Addr     string
	Username string
	Password string
	Key      string
	Timeout  time.Duration
}

type TimeoutConn struct {
	net.Conn
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (c *TimeoutConn) Read(b []byte) (int, error) {
	err := c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
	if err != nil {
		return 0, err
	}
	return c.Conn.Read(b)
}

func (c *TimeoutConn) Write(b []byte) (int, error) {
	err := c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	if err != nil {
		return 0, err
	}
	return c.Conn.Write(b)
}

func NewSSHClient() *SSHClient {
	return &SSHClient{
		initClient: &sync.Once{},
	}
}

func (c *SSHClient) buildSSHConfig() (*ssh.ClientConfig, error) {
	sshConfig := &ssh.ClientConfig{
		User:            c.Username,
		Timeout:         c.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if c.Password != "" {
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(c.Password)}
	}
	if c.Key != "" {
		var data []byte
		var err error
		if strings.Index(c.Key, "@") == 0 {
			data, err = ioutil.ReadFile(string(c.Key[1:]))
			if err != nil {
				return nil, err
			}
		} else {
			data = []byte(c.Key)
		}

		signer, err := ssh.ParsePrivateKey(data)
		if err != nil {
			return nil, err
		}

		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	}
	return sshConfig, nil
}

func (c *SSHClient) Reconnect() error {
	fmt.Println("reconnect")
	c.initClient = &sync.Once{}
	return c.Connect()
}

func (c *SSHClient) Connect() error {

	c.initClient.Do(func() {

		sshConfig, err := c.buildSSHConfig()
		if err != nil {
			c.sshError = err
			return
		}
		conn, err := net.DialTimeout("tcp", c.Addr, c.Timeout)
		if err != nil {
			c.sshError = err
			return
		}
		TimeoutConn := &TimeoutConn{conn, c.Timeout, c.Timeout}
		cli, chans, reqs, err := ssh.NewClientConn(TimeoutConn, c.Addr, sshConfig)
		if err != nil {
			c.sshError = err
			return
		}
		client := ssh.NewClient(cli, chans, reqs)
		c.sshclient = client
		c.sshError = nil
		go func() {
			t := time.NewTicker(3 * time.Second)
			defer t.Stop()
			for {
				<-t.C
				_, _, err := client.Conn.SendRequest("keepalive@epicon.com", true, nil)
				if err != nil {
					return
				}
			}
		}()
	})
	return c.sshError
}

func (c *SSHClient) Close() {
	c.sshclient.Close()
}

func (c *SSHClient) Run(dir, command string, args ...string) (string, error) {
	err := c.Connect()
	if err != nil {
		err = c.Reconnect() // always Reconnect if it's failed in first connect
		if err != nil {
			return "", err
		}
	}
	// start := time.Now()
	session, err := c.sshclient.NewSession()
	if err != nil {
		c.sshError = err
		return "", err
	}
	if err != nil {
		c.sshError = err
		return "", err
	}

	var stdoutB bytes.Buffer
	session.Stdout = &stdoutB
	var stderrB bytes.Buffer
	session.Stderr = &stderrB

	err = session.Run(util.BuildCmd(dir, command, args...))

	defer session.Close()
	if err != nil {
		c.sshError = err
		exitErr := &ssh.ExitMissingError{}
		if err.Error() == exitErr.Error() {
			return "", fmt.Errorf("Connection is disconnected by the Timeout or lost")
		}
		return "", fmt.Errorf("%v %v", string(stderrB.Bytes()), err.Error())
	}
	// fmt.Println(time.Since(start))
	if stderrB.Len() == 0 {
		c.sshError = nil
		err = nil
	} else {
		err = fmt.Errorf("%v", string(stderrB.Bytes()))
		c.sshError = err
	}
	return strings.Trim(string(stdoutB.Bytes()), "\n"), err

}

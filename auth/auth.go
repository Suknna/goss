package auth

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type Opts struct {
	IP             string
	Port           string
	RemoteUser     string
	RemotePassword string
	TimeOut        int
	Auto           bool
}

// 用于完成ssh认证操作

func Init(opts *Opts) (*ssh.Client, error) {
	sshClient, err := ssh.Dial("tcp", net.JoinHostPort(opts.IP, opts.Port), &ssh.ClientConfig{
		User: opts.RemoteUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(opts.RemotePassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: time.Second * time.Duration(opts.TimeOut),
	})
	if err != nil {
		return nil, fmt.Errorf("%s@%s pass: %s connect failed, %s", opts.RemoteUser, opts.IP, opts.RemotePassword, err)
	}
	return sshClient, nil
}

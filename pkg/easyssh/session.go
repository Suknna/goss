package easyssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type CtxSession struct {
	*ssh.Session
	ctx context.Context
}

func NewCTXSession(ctx context.Context, con *ssh.Client) (*CtxSession, error) {
	ses, err := con.NewSession()
	return &CtxSession{
		Session: ses,
		ctx:     ctx,
	}, err
}

// 带有超时的命令执行控制
func (session *CtxSession) Execute(cmd string) (string, error) {
	defer func() {
		if session != nil {
			session.Close()
		}
	}()
	errCH := make(chan error, 1)
	var OutPutResult string
	go func() {
		defer close(errCH)
		out, err := session.CombinedOutput(cmd)
		if err != nil {
			errCH <- errors.New(string(out))
			return
		}
		slog.Debug("命令执行完成",
			"out", out)
		OutPutResult = string(out)
	}()

	for {
		select {
		case <-session.ctx.Done():
			// 执行信号断开
			session.Signal(ssh.SIGTERM)
			time.Sleep(100 * time.Millisecond) // 信号处理时间
			return OutPutResult, fmt.Errorf("timeout")
		case err := <-errCH:
			return OutPutResult, err
		}
	}
}

func (session *CtxSession) ExecutePrivilegedCommandOverSSH(command string, sudoPassword string) (string, error) {
	defer func() {
		if session != nil {
			session.Close()
		}
	}()
	mode := ssh.TerminalModes{
		ssh.ECHO:          0,     // 禁用回显，防止密码泄露
		ssh.TTY_OP_ISPEED: 14400, // 输入速度限制
		ssh.TTY_OP_OSPEED: 14400, // 输出速度限制
	}
	// 配置终端,默认使用linux
	if err := session.RequestPty("linux", 80, 24, mode); err != nil {
		return "", fmt.Errorf("failed to remotely request pty of Linux type, %w", err)
	}
	session.Setenv("LANG", "C")
	stdIn, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("get remote standard input exception, %w", err)
	}
	errChan := make(chan error, 1)
	stdOutBuf := new(bytes.Buffer)
	session.Stdout = stdOutBuf
	go func(in io.WriteCloser, out *bytes.Buffer) {
		for {
			if strings.Contains(out.String(), "Password") {
				// 注入root密码
				_, err := in.Write([]byte(sudoPassword + "\n"))
				if err != nil {
					errChan <- err
					break
				}
				break
			}
		}
	}(stdIn, stdOutBuf)
	go func() {
		defer close(errChan)
		errChan <- session.Run(fmt.Sprintf("env LANG=C  su - root -c \"%s\"", command))
	}()
	// 拼接命令
	for {
		select {
		case <-session.ctx.Done():
			// 执行信号断开
			session.Signal(ssh.SIGTERM)
			time.Sleep(100 * time.Millisecond) // 信号处理时间
			return cleanOutput(stdOutBuf.String()), fmt.Errorf("timeout")
		case err := <-errChan:
			return cleanOutput(stdOutBuf.String()), err
		}
	}
}

func cleanOutput(output string) string {
	passwordPrompts := "Password:"

	lines := strings.Split(output, "\n")
	cleanLines := []string{}
	for _, line := range lines {
		isPrompt := false
		if strings.Contains(line, passwordPrompts) {
			isPrompt = true
			continue
		}
		if !isPrompt {
			cleanLines = append(cleanLines, line)
		}
	}
	return strings.Join(cleanLines, "\n")
}

package execute

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

func SudoExecute(password string, command string, sshClient *ssh.Client) (string, error) {
	// 创建会话
	session, err := sshClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("create session failed, err: %s", err.Error())
	}
	// 执行命令
	out, err := run(fmt.Sprintf(`echo %s | su - root -c "%s"`, password, command), session)
	if err != nil {
		return "", fmt.Errorf("execute command failed. %s", out)
	}
	return out, nil
}

func Execute(command string, sshClient *ssh.Client) (string, error) {
	// 创建client连接
	session, err := sshClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("create session failed, err: %s", err.Error())
	}
	// 执行命令
	out, err := run(command, session)
	if err != nil {
		return "", fmt.Errorf("execute command failed. %s", out)
	}
	// 去除最后一个回车
	return out, nil
}

func run(cmd string, session *ssh.Session) (string, error) {
	var output []string
	b, err := session.CombinedOutput(cmd)
	lastIndex := bytes.LastIndexByte(b, '\n')
	if lastIndex == -1 {
		return string(b), err
	}
	line := string(append(b[:lastIndex], b[lastIndex+1:]...))
	for i, l := range strings.Split(line, "\n") {
		if i == 0 {
			if strings.HasPrefix(strings.TrimSpace(l), "Password") {
				output = append(output, strings.ReplaceAll(l, "Password: ", ""))
			} else if strings.HasPrefix(strings.TrimSpace(l), "密码") {
				output = append(output, strings.ReplaceAll(l, "密码: ", ""))
			} else {
				output = append(output, l)
			}
		} else {
			output = append(output, l)
		}
	}
	return strings.Join(output, "\n"), err
}

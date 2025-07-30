package easyssh

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SecurityMode int

const (
	// SecurityModePermissive 自动接受所有密钥（仅限调试环境）
	SecurityModePermissive SecurityMode = iota // 0

	// SecurityModeInteractive 要求用户确认未知/冲突密钥（超时默认拒绝）
	SecurityModeInteractive // 1

	// SecurityModeStrict 仅信任本地已知密钥（生产环境推荐）
	SecurityModeStrict // 2
)

type Opts struct {
	IP string
	// default: 22
	Port   string
	User   string
	Passwd string
	// 连接阶段超时(秒)
	ConnectTimeout int
	// default: SecurityModeInteractive
	Mode SecurityMode
}

func NewClient(opts Opts) (*ssh.Client, error) {
	//创建ssh连接
	return ssh.Dial("tcp", net.JoinHostPort(opts.IP, opts.Port), &ssh.ClientConfig{
		User: opts.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(opts.Passwd),
			ssh.KeyboardInteractive(func(name, instruction string, questions []string, echos []bool) (answers []string, err error) {
				answers = make([]string, len(questions))
				for i, q := range questions {
					qLower := strings.ToLower(q)
					switch {
					case strings.Contains(qLower, "passwd") && !echos[i]:
						answers[i] = opts.Passwd
					default:
						answers[i] = ""
					}
				}
				return answers, nil
			}),
		},
		HostKeyCallback: keyProcessing(opts.Mode),
		Timeout:         time.Second * time.Duration(opts.ConnectTimeout),
	})
}

func keyProcessing(mode SecurityMode) func(hostname string, remote net.Addr, key ssh.PublicKey) error {
	switch mode {
	case SecurityModePermissive:
		return ssh.InsecureIgnoreHostKey()
	case SecurityModeInteractive:
		return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// 获取公钥指纹文件路径
			f, path, err := getPublicKeyFingerprintFile()
			if err != nil {
				return err
			}
			// 比对指纹
			callback, err := knownhosts.New(path)
			if err != nil {
				return err
			}
			err = callback(hostname, remote, key)
			if err == nil {
				// 密钥已知
				return nil
			}
			if keyErr, ok := err.(*knownhosts.KeyError); ok {
				if len(keyErr.Want) > 0 {
					slog.Warn(fmt.Sprintf("Host: %s key mismatch!", hostname))
					for _, knownKey := range keyErr.Want {
						fmt.Printf("- %s\n", ssh.FingerprintSHA256(knownKey.Key))
					}
				}
			} else if revokedErr, ok := err.(*knownhosts.RevokedError); ok {
				return fmt.Errorf("connection refused: Host %s  key has been revoked (%s)", hostname, ssh.FingerprintSHA256(revokedErr.Revoked.Key))
			} else {
				// 其他错误
				return fmt.Errorf("failed to verify host key: %w", err)
			}
			// 密钥不存在或不匹配，提示用户
			fmt.Printf("New host %s key fingerprint: %s\nAccept and save? (y/n):", hostname, ssh.FingerprintSHA256(key))
			var output string
			for {
				fmt.Scanln(output)
				if output != "y" && output != "n" {
					fmt.Print("Please enter 'y' or 'n' : ")
					time.Sleep(time.Second * 1)
					continue
				}
				break
			}
			if output == "y" {
				// 读取公钥指纹文件
				bs, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("failed to read known_hosts file: %w", err)
				}
				var newLines []string
				scanner := bufio.NewScanner(bytes.NewReader(bs))
				for scanner.Scan() {
					line := scanner.Text()
					if strings.HasPrefix(line, hostname+",") || strings.HasPrefix(line, "|1|") {
						continue
					}
					newLines = append(newLines, line)
				}
				// 写入新内容（先删除旧条目）
				if err := os.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), f.Mode()); err != nil {
					return fmt.Errorf("failed to update the known_hosts file: %w", err)
				}
			}
			file, err := os.OpenFile(path,
				os.O_APPEND|os.O_WRONLY|os.O_CREATE, f.Mode())
			if err != nil {
				return fmt.Errorf("unable to open the known_hosts file: %w", err)
			}
			defer file.Close()

			_, err = file.WriteString(knownhosts.Line([]string{hostname}, key) + "\n")
			return err
		}
	case SecurityModeStrict:
		return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			_, path, err := getPublicKeyFingerprintFile()
			if err != nil {
				return err
			}
			callback, err := knownhosts.New(os.ExpandEnv(path))
			if err != nil {
				return fmt.Errorf("failed to load the known_hosts file: %w", err)
			}
			return callback(hostname, remote, key)
		}

	default:
		return ssh.InsecureIgnoreHostKey()
	}
}

// 获取公钥指纹存储目录
func getPublicKeyFingerprintFile() (os.FileInfo, string, error) {
	// 获取用户目录
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve user home directory: %w", err)
	}
	// 判断.ssh是否存在
	fs, err := os.Stat(path.Join(userHomeDir, ".ssh"))
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path.Join(userHomeDir, ".ssh"), 0700); err != nil {
				return nil, "", fmt.Errorf("failed to create .ssh directory: %w", err)
			}
		}
		return nil, "", fmt.Errorf("unable to access .ssh directory: %w", err)
	}
	publicKeyFingerprintPath := path.Join(userHomeDir, ".ssh/known_hosts")
	// 判断公钥指纹文件是否存在
	_, err = os.Stat(publicKeyFingerprintPath)
	if err != nil {
		// 判断是否因为不存在导致的错误
		if os.IsNotExist(err) {
			f, err := os.Create(publicKeyFingerprintPath)
			if err != nil {
				return nil, "", fmt.Errorf("failed to create the known_hosts file: %w", err)
			}
			f.Close()
		}
		return nil, "", fmt.Errorf("unable to access the known_hosts file: %w", err)
	}
	return fs, publicKeyFingerprintPath, nil
}

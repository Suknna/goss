package auth

/*
配置密钥的认证方式
*/
import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func VerifyHost(host string, remote net.Addr, key ssh.PublicKey) error {
	// 检查hostkey是否存在
	hostFound, err := checkKnownHost(host, remote, key, "")
	if hostFound && err != nil {
		slog.Warn(err.Error())
		return nil
	}
	if hostFound && err == nil {
		return nil
	}
	if !askIsHostTrusted(host, key) {
		return errors.New("you typed no, aborted")
	}
	return addKnownHost(host, remote, key, "")
}

func knownHosts(file string) (ssh.HostKeyCallback, error) {
	return knownhosts.New(file)
}

// 默认方式读取本地know_hosts文件内容
// func defaultUserKnownHosts() (ssh.HostKeyCallback, error) {
// 	// 获取用户家目录
// 	home, err := defaultUserKnownHostsPath()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return knownHosts(home)
// }

// 检测hostskey
func checkKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string) (found bool, err error) {

	var keyErr *knownhosts.KeyError

	// 默认采用用户自己的knows_hosts
	if knownFile == "" {
		path, err := defaultUserKnownHostsPath()
		if err != nil {
			return false, err
		}

		knownFile = path
	}

	// 从本地保存的know_hosts文件中读取信息
	callback, err := knownHosts(knownFile)
	if err != nil {
		return false, err
	}

	// 根据远端返回的key信息和本地knowhosts中返回的进行对比，err为空表示key信息对比正确
	err = callback(host, remote, key)
	if err == nil {
		return true, nil
	}

	// 解析error的内容，如果找到对应主机和对应的key但是key不一致也返回正常找到，但是key的信息不对
	if errors.As(err, &keyErr) && len(keyErr.Want) > 0 {
		return true, keyErr
	}
	// 上述两种情况都不对，表示没找到主机和对应的key
	return false, err
}

// 添加主机和对应的key解析
func addKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string) (err error) {

	// 使用用户默认的know_hosts路径
	if knownFile == "" {
		path, err := defaultUserKnownHostsPath()
		if err != nil {
			return err
		}

		knownFile = path
	}

	f, err := os.OpenFile(knownFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	remoteNormalized := knownhosts.Normalize(remote.String())
	hostNormalized := knownhosts.Normalize(host)
	addresses := []string{remoteNormalized}

	if hostNormalized != remoteNormalized {
		addresses = append(addresses, hostNormalized)
	}

	_, err = f.WriteString(knownhosts.Line(addresses, key) + "\n")

	return err
}

func defaultUserKnownHostsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.ssh/known_hosts", home), nil
}

func askIsHostTrusted(host string, key ssh.PublicKey) bool {

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Unknown Host: %s \nFingerprint: %s \n", host, ssh.FingerprintSHA256(key))
	fmt.Print("Would you like to add it? type yes or no: ")

	a, err := reader.ReadString('\n')

	if err != nil {
		log.Fatal(err)
	}

	return strings.ToLower(strings.TrimSpace(a)) == "yes"
}

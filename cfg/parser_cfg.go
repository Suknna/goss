package cfg

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Goss struct {
	Global Global   `yaml:"global"`
	Groups []*Group `yaml:"groups"`
	Runs   []*Run   `yaml:"runs"`
}

type Global struct {
	TimeOut      int    `yaml:"timeOut"`
	WorkLoad     int    `yaml:"workLoad"`
	RemoteUser   string `yaml:"remoteUser"`
	RemotePasswd string `yaml:"remotePasswd"`
	Auto         bool   `yaml:"auto"`
}

type Group struct {
	Name         string   `yaml:"name"`
	RemoteUser   string   `yaml:"remoteuser"`
	RemotePasswd string   `yaml:"remotepasswd"`
	Address      []string `yaml:"address"`
	SudoPass     []string `yaml:"sudopass"`
}

type Run struct {
	Name     string    `yaml:"name"`
	Groups   []string  `yaml:"groups"`
	Upload   *Upload   `yaml:"upload"`
	Download *Download `yaml:"download"`
	Execute  *Execute  `yaml:"execute"`
	Ignore   bool      `yaml:"ignore"`
}

type Upload struct {
	RemotePath string `yaml:"remotePath"`
	LocalPath  string `yaml:"localPath"`
	Cover      bool   `yaml:"cover"` // 是否覆盖远端重名文件
}

type Download struct {
	RemotePath string `yaml:"remotePath"`
	LocalPath  string `yaml:"localPath"`
	Cover      bool   `yaml:"cover"` // 是否覆盖本地重名文件，如果不覆盖则自动重命名
}

type Execute struct {
	Command string `yaml:"command"`
}

func Parser(path string) (*Goss, error) {
	viper.SetConfigFile(path) // Set the path to your config file here
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("fatal error config file: %w", err)
	}

	var goss Goss
	err = viper.Unmarshal(&goss)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}
	// 验证配置
	if err := configVerify(&goss); err != nil {
		return nil, fmt.Errorf("fatal error config file: %s", err)
	}
	return &goss, nil
}

func ParserIP(address string) (ip, port string) {
	// 去除多余空格
	address = strings.TrimSpace(address)
	// 判断是否有:
	if strings.Contains(address, ":") {
		// 切割返回
		ip = strings.Split(address, ":")[0]
		port = strings.Split(address, ":")[1]
		return ip, port
	}
	return address, "22"
}

func configVerify(o *Goss) error {
	if o.Global.TimeOut <= 0 {
		return fmt.Errorf("global.timeOut value is less than or equal to 0")
	} else if o.Global.WorkLoad <= 0 {
		return fmt.Errorf("global.workLoad value is less than or equal to 0")
	}
	if o.Runs == nil {
		return fmt.Errorf("no 'runs' information detected, please confirm if the variable name is correct")
	}
	if o.Groups == nil {
		return fmt.Errorf("no 'groups' information detected, please confirm if the variable name is correct")
	}
	for _, group := range o.Groups {
		if group.RemoteUser == "" && group.RemotePasswd != "" || group.RemotePasswd == "" && group.RemoteUser != "" {
			return fmt.Errorf("groups.name = '%s'. If you want to create remote connection users and passwords for a certain group, please ensure that both the username and password are not empty", group.Name)
		}
		if len(group.Address) == 0 {
			return fmt.Errorf("groups.name = '%s'. no valid address detected", group.Name)
		}
	}
	return nil
}

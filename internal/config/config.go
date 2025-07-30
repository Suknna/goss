package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type FileTransferPolicy string

const (
	Always FileTransferPolicy = "always"
	Never  FileTransferPolicy = "never"
)

// 默认值常量定义
const (
	// Connection 默认值
	DefaultPort           = 22
	DefaultConnectTimeout = 3
	DefaultSecurityMode   = 0

	// Execution 默认值
	DefaultMaxWorkers  = 1
	DefaultTaskTimeout = 120

	// FileTransfer 默认值
	DefaultUploadDir       = "/tmp"
	DefaultDownloadDir     = "./downloads/"
	DefaultOverwritePolicy = Always
	DefaultTransferTimeout = 1800
	DefaultRetries         = 3
)

// 全局配置结构体
type GossConfig struct {
	Connection   *ConnectionConfig   `mapstructure:"connection"`
	Execution    *ExecutionConfig    `mapstructure:"execution"`
	FileTransfer *FileTransferConfig `mapstructure:"file_transfer"`
}

type ConnectionConfig struct {
	DefaultPort    int `mapstructure:"default_port"`
	ConnectTimeout int `mapstructure:"connect_timeout"`
	SecurityMode   int `mapstructure:"security_mode"`
}

type ExecutionConfig struct {
	MaxWorkers  int  `mapstructure:"max_workers"`
	TaskTimeout int  `mapstructure:"task_timeout"`
	StopOnError bool `mapstructure:"stop_on_error"`
}

type FileTransferConfig struct {
	DefaultUploadDir   string             `mapstructure:"default_upload_dir"`
	DefaultDownloadDir string             `mapstructure:"default_download_dir"`
	OverwritePolicy    FileTransferPolicy `mapstructure:"overwrite_policy"`
	TransferTimeout    int                `mapstructure:"transfer_timeout"`
	Retries            int                `mapstructure:"retries"`
}

// LoadConfig 加载全局配置
func LoadConfig(path string) (*GossConfig, error) {
	v := viper.New()
	// 设置默认值
	v.SetDefault("connection.default_port", DefaultPort)
	v.SetDefault("connection.connect_timeout", DefaultConnectTimeout)
	v.SetDefault("connection.security_mode", DefaultSecurityMode)
	v.SetDefault("execution.max_workers", DefaultMaxWorkers)
	v.SetDefault("execution.task_timeout", DefaultTaskTimeout)
	v.SetDefault("file_transfer.default_upload_dir", DefaultUploadDir)
	v.SetDefault("file_transfer.default_download_dir", DefaultDownloadDir)
	v.SetDefault("file_transfer.overwrite_policy", DefaultOverwritePolicy)
	v.SetDefault("file_transfer.transfer_timeout", DefaultTransferTimeout)
	v.SetDefault("file_transfer.retries", DefaultRetries)

	v.SetConfigType("yaml")
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read goss configuration file: %s", err.Error())
		}
	}

	var cfg GossConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse goss configuration: %s", err.Error())
	}
	if err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("goss configuration validation failed: %s", err.Error())
	}
	return &cfg, nil
}

// validateConfig 配置验证
func ValidateConfig(cfg *GossConfig) error {

	if cfg.Connection.DefaultPort <= 0 {
		return fmt.Errorf("default_port must be greater than 0")
	}

	if cfg.Connection.ConnectTimeout <= 0 {
		return fmt.Errorf("connect_timeout must be greater than 0")
	}

	if cfg.Connection.SecurityMode < 0 || cfg.Connection.SecurityMode > 2 {
		return fmt.Errorf("security_mode must be between 0 and 2")
	}
	if cfg.Execution.MaxWorkers <= 0 {
		return fmt.Errorf("max_workers must be greater than 0")
	}

	if cfg.Execution.TaskTimeout <= 0 {
		return fmt.Errorf("task_timeout must be greater than 0")
	}

	switch cfg.FileTransfer.OverwritePolicy {
	case Always, Never:
		// 有效值，不做处理
	default:
		return fmt.Errorf("overwrite_policy must be one of: always, never")
	}

	if cfg.FileTransfer.TransferTimeout <= 0 {
		return fmt.Errorf("transfer_timeout must be greater than 0")
	}

	if cfg.FileTransfer.Retries <= 0 {
		return fmt.Errorf("retries must be greater than 0")
	}

	return nil
}

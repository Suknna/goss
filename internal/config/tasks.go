package config

import (
	"fmt"
	"goss/internal/utils"
	"log/slog"

	"github.com/spf13/viper"
)

type TaskType string

const (
	CMD      TaskType = "cmd"
	SCRIPT   TaskType = "script"
	UPLOAD   TaskType = "upload"
	DOWNLOAD TaskType = "download"
)

type Task struct {
	Type        TaskType `mapstructure:"type"`
	Description string   `mapstructure:"description"`
	Cmd         string   `mapstructure:"cmd"`
	RequireSudo bool     `mapstructure:"require_sudo"`
	Local       string   `mapstructure:"local"`
	Remote      string   `mapstructure:"remote"`
}

// LoadTasks 加载任务配置
func LoadTasks(configPath string) ([]*Task, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read task configuration: %w", err)
	}

	var tasks []*Task
	if err := v.UnmarshalKey("tasks", &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse task configuration: %w", err)
	}

	// 调用纠错框架
	if err := ValidateTasks(tasks); err != nil {
		return nil, fmt.Errorf("task configuration validation failed: %s", err.Error())
	}

	return tasks, nil
}

// validateTasks 任务验证
func ValidateTasks(tasks []*Task) error {
	for i, task := range tasks {
		if task.Local != "" {
			_, err := utils.RenderPathTemplate(task.Local, "127.0.0.1")
			if err != nil {
				return fmt.Errorf("template rendering failed local %s %s. Index %d", task.Local, err, i)
			}
		}

		// 处理Remote路径
		if task.Remote != "" {
			_, err := utils.RenderPathTemplate(task.Remote, "127.0.0.1")
			if err != nil {
				return fmt.Errorf("template rendering failed remote %s %s. Index %d", task.Remote, err, i)
			}
		}
		if task.Description == "" {
			return fmt.Errorf("the task description is mandatory. Index %d", i+1)
		}
		switch task.Type {
		case CMD:
			if task.Cmd == "" {
				return fmt.Errorf("the 'cmd' parameter of the command execution task cannot be empty. Index %d", i+1)
			}
		case SCRIPT:
			if task.Cmd == "" {
				return fmt.Errorf("the 'cmd' parameter for the script execution task cannot be empty. Index %d", i+1)
			}
			if task.Local == "" {
				return fmt.Errorf("the 'local' parameter of the script execution task cannot be empty. Index %d", i+1)
			}
			if task.Remote == "" {
				slog.Warn("The 'remote' parameter for the script execution task is empty, the default path will be used.", slog.String("remote", DefaultUploadDir))
				task.Remote = DefaultUploadDir
			}
		case UPLOAD:
			if task.Remote == "" {
				slog.Warn("The 'remote' parameter for the upload task is empty, the default path will be used.", slog.String("remote", DefaultUploadDir))
				task.Remote = DefaultUploadDir
			}
			if task.Local == "" {
				return fmt.Errorf("the 'local' parameter of the upload task cannot be empty. Index %d", i+1)
			}
		case DOWNLOAD:
			if task.Local == "" {
				slog.Warn("The 'local' parameter for the download task is empty, the default path will be used.", slog.String("local", DefaultDownloadDir))
				task.Remote = DefaultUploadDir
			}
			if task.Remote == "" {
				return fmt.Errorf("the 'remote' parameter of the download task cannot be empty. Index %d", i+1)
			}
		default:
			return fmt.Errorf("unknown task type. Index %d", i+1)
		}
	}
	return nil
}

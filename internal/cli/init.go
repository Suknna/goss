/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate configuration file template.",
	Long:  `Generate configuration file templates for hosts.ini and tasks.yml in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		templatePath := "./play-" + time.Now().Format("2006-01-02T150405")
		hostsPath := filepath.Join(templatePath, "hosts.ini")
		tasksPath := filepath.Join(templatePath, "tasks.yml")
		gossPath := filepath.Join(templatePath, "goss_config.yaml")
		// 创建配置存放目录
		if err := os.MkdirAll(templatePath, 0777); err != nil {
			fmt.Printf("Failed to create the configuration storage directory. %s\n", err)
			return
		}
		// 生成hosts.ini模板
		if err := os.WriteFile(hostsPath, []byte(hostsTemplate), 0644); err != nil {
			fmt.Printf("Failed to generate hosts.ini %s\n", err)
			return
		}

		// 生成tasks.yml模板
		if err := os.WriteFile(tasksPath, []byte(tasksTemplate), 0644); err != nil {
			fmt.Printf("Failed to generate tasks.yml %s\n", err)
			return
		}
		// 生成goss工具全局配置
		if err := os.WriteFile(gossPath, []byte(goss_configTemplate), 0644); err != nil {
			fmt.Printf("Failed to generate goss_config.ini %s\n", err)
			return
		}
		fmt.Printf("Configuration file templates have been generated:  \n%s, \n%s, \n%s\n", hostsPath, tasksPath, gossPath)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

const hostsTemplate = `# 格式：IP,用户名,登录密码,特权密码
# 示例：
# 192.168.1.101,admin,P@ssw0rd123,SudoP@ss!
# 10.0.5.17,deploy,Deploy123,Root!789
# 172.16.0.33,ubuntu,UbuntuPass,  # 无特权账户留空
`

const tasksTemplate = `# tasks.yaml
# ============ 任务定义 ============
# 支持四种任务类型：cmd, script, upload, download
# 执行时按顺序执行 tasks 列表中的任务
# 路径相关配置支持变量注入：{{ .IP }}或者{{ .TIME }} 
## 例如: 路径./download/{{ .IP }}_{{ .TIME }}.txt程序会自动格式化最终展示为: ./download/192.168.200.2_20250403.txt

#tasks:
#  # 1. 命令执行任务
#  - type: cmd
#    description: "检查磁盘空间"  # 任务描述
#    cmd: "df -h | grep -v tmpfs"  # 实际执行的命令
#    require_sudo: false           # 是否使用特权用户执行
#    
#  # 2. 脚本执行任务
#  - type: script
#    description: "部署应用"
#    local: "/local/scripts/deploy.sh"  # 本地脚本路径
#    remote: "/opt/scripts/"      # 上传到远程的目录（可选，默认用file_transfer.default_upload_dir）
#    cmd: "sh /local/scripts/deploy.sh"  # 脚本执行命令
#    require_sudo: true           # 是否特权执行
#    
#  # 3. 文件上传任务
#  - type: upload
#    description: "上传配置文件"
#    local: "./configs/app.conf"  # 本地文件路径 
#    remote: "/etc/app.conf"      # 远程路径（可选）
#    
#  # 4. 文件下载任务
#  - type: download
#    description: "下载日志文件"
#    remote: "/var/log/nginx/*.log"  # 远程目录
#    local: "./logs/"         # 本地存储目录（可选）
#
`
const goss_configTemplate = `connection:
  default_port: 22
  connect_timeout: 3
  security_mode: 0

execution:
  max_workers: 1
  task_timeout: 120
  stop_on_error: true

file_transfer:
  default_upload_dir: "/tmp"
  default_download_dir: "./downloads/"
  overwrite_policy: "always"
  transfer_timeout: 1800
  retries: 3`

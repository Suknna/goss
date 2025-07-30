/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"fmt"

	"goss/internal/config"
	"goss/internal/dispatcher"

	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a single task.",
	Long: `Execute tasks on a single host, inheriting all parameter configurations of the root command
and supporting direct specification of task parameters.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 构建内存中的任务配置
		tasks, err := buildTaskConfig(cmd)
		if err != nil {
			fmt.Printf("configuration error %s\n", err)
			return
		}
		// 加载全局配置
		hosts, cfg, err := basicConfigurationParserconfigParser(HostPath)
		if err != nil {
			fmt.Println(err)
			return
		}
		dispatcher.Run(hosts, tasks, cfg, Save)
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	// Task execution parameters
	execCmd.Flags().String("type", "", "Command execution types: script, cmd, download, upload")
	execCmd.Flags().String("cmd", "", "Command string (required for 'cmd' type)")
	execCmd.Flags().String("local", "", "Local file/directory path (required for 'upload'/'download'/'script' types)")
	execCmd.Flags().String("remote", "", "Remote file/directory path (required for 'upload'/'download'/'script' types)")
	execCmd.Flags().Bool("sudo", false, "Require sudo privileges for execution")
	// 必须条件配置

}

// Build task configuration
func buildTaskConfig(cmd *cobra.Command) ([]*config.Task, error) {
	// Get task type
	taskType, err := cmd.Flags().GetString("type")
	if err != nil {
		return nil, fmt.Errorf("failed to get task type: %s", err.Error())
	}
	taskCmd, err := cmd.Flags().GetString("cmd")
	if err != nil {
		return nil, fmt.Errorf("failed to get task cmd: %s", err.Error())
	}
	taskLocal, err := cmd.Flags().GetString("local")
	if err != nil {
		return nil, fmt.Errorf("failed to get task local: %s", err.Error())
	}
	taskRemote, err := cmd.Flags().GetString("remote")
	if err != nil {
		return nil, fmt.Errorf("failed to get task remote: %s", err.Error())
	}
	taskSudo, err := cmd.Flags().GetBool("sudo")
	if err != nil {
		return nil, fmt.Errorf("failed to get task sudo: %s", err.Error())
	}
	return []*config.Task{
		{
			Type:        config.TaskType(taskType),
			Description: "Single execution",
			Cmd:         taskCmd,
			RequireSudo: taskSudo,
			Local:       taskLocal,
			Remote:      taskRemote,
		},
	}, nil
}

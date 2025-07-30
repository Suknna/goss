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

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Execute tasks in the configuration file.",
	Long:  `Execute batch tasks based on the configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		taskPath, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Println(err)
			return
		}
		// 加载全局配置
		hosts, cfg, err := basicConfigurationParserconfigParser(HostPath)
		if err != nil {
			fmt.Println(err)
			return
		}
		// 加载配置
		tasks, err := config.LoadTasks(taskPath)
		if err != nil {
			fmt.Printf("Failed to parse task configuration %s\n", err.Error())
			return
		}
		//执行任务逻辑
		dispatcher.Run(hosts, tasks, cfg, Save)
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringP("file", "f", "./tasks.yml", "Task configuration file path")
	applyCmd.MarkFlagRequired("file")
}

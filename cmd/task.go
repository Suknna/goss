/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"goss/cfg"
	"goss/engine"
	"os"

	"github.com/spf13/cobra"
)

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Run tasks according to the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		filePath, _ := cmd.Flags().GetString("file")
		// 解析配置文件
		goss, err := cfg.Parser(filePath)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		engine.AutomaticExecution(goss)
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// taskCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// taskCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	taskCmd.Flags().StringP("file", "F", "./task.yml", "Develop task configuration file path, please use explain to view configuration details.")
}

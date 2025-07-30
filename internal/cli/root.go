/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"goss/internal/config"
	"os"

	"github.com/spf13/cobra"
)

var (
	Save       string
	HostPath   string
	ConfigPath string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "goss",
	// Short: "",
	Long: `A lightweight SSH/SFTP-based operations tool.

	Find more information at: https://github.com/Suknna/goss

	`,
	Version: "V0.2.0",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&Save, "save", "", "The output format supports (json, excel).")
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "goss_config.yaml", "Specify the location of goss environment variables. A template configuration can be generated using the init subcommand.")
	rootCmd.PersistentFlags().StringVar(&HostPath, "hosts", "hosts.ini", "Host information configuration file path")
}

// 基础配置解析器负责解析cli全局配置和主机信息配置
func basicConfigurationParserconfigParser(HostPath string) ([]*config.Host, *config.GossConfig, error) {
	// 解析hosts配置
	hosts, err := config.ParseHostsFile(HostPath)
	if err != nil {
		return nil, nil, err
	}
	// 解析全局配置
	cfg, err := config.LoadConfig(ConfigPath)
	if err != nil {
		return nil, nil, err
	}
	return hosts, cfg, nil
}

/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"goss/auth"
	"goss/cfg"
	"goss/engine"
	"goss/record"

	"github.com/spf13/cobra"
)

// executeCmd represents the execute command
var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Remote execution of commands or scripts",
	Run: func(cmd *cobra.Command, args []string) {
		command, _ := cmd.Flags().GetString("command")
		ip, port := cfg.ParserIP(address)
		record.SSHRecord(ip, "execute", func(args ...interface{}) (string, error) {
			return engine.Execute(&auth.Opts{
				IP:             ip,
				Port:           port,
				RemoteUser:     remoteUser,
				RemotePassword: remotePasswd,
				TimeOut:        timeout,
			}, sudoPass, command)
		})
	},
}

func init() {
	rootCmd.AddCommand(executeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// executeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// executeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	executeCmd.Flags().StringP("command", "", "", "Remote command execution.")
}

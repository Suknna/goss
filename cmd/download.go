/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"goss/auth"
	"goss/cfg"
	"goss/engine"
	"goss/record"

	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download files from remote servers.",
	Run: func(cmd *cobra.Command, args []string) {
		remotePath, _ := cmd.Flags().GetString("remotePath")
		localPath, _ := cmd.Flags().GetString("localPath")
		isCover, _ := cmd.Flags().GetBool("cover")
		ip, port := cfg.ParserIP(address)
		if isCover {
			localPath = fmt.Sprintf("%s_%s", remotePath, ip)
		}
		record.SftpTaskRecord(ip, "download", func(args ...interface{}) error {
			return engine.Download(&auth.Opts{
				IP:             ip,
				Port:           port,
				RemoteUser:     remoteUser,
				RemotePassword: remotePasswd,
				TimeOut:        timeout,
				Auto:           auto,
			}, remotePath, localPath)
		})
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	downloadCmd.Flags().StringP("remotePath", "r", "", "The path of the remote server.")
	downloadCmd.Flags().StringP("localPath", "l", "", "The path of the local server.")
	downloadCmd.Flags().BoolP("cover", "", false, "Is the local file or directory with the same name overwritten. If not covered, create and save according to the file name and IP address. For example: info.log_192.168.200.2")
}

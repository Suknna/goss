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

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload files to remote servers.",
	Run: func(cmd *cobra.Command, args []string) {
		remotePath, _ := cmd.Flags().GetString("remotePath")
		localPath, _ := cmd.Flags().GetString("localPath")
		isCover, _ := cmd.Flags().GetBool("cover")
		ip, port := cfg.ParserIP(address)
		if isCover {
			remotePath = fmt.Sprintf("%s_%s", localPath, ip)
		}
		record.SftpTaskRecord(ip, "upload", func(args ...interface{}) error {
			return engine.Upload(&auth.Opts{
				IP:             ip,
				Port:           port,
				RemoteUser:     remoteUser,
				RemotePassword: remotePasswd,
				TimeOut:        int(timeout),
				Auto:           auto,
			}, remotePath, localPath)
		})
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uploadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uploadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	uploadCmd.Flags().StringP("remotePath", "r", "", "The path of the remote server.")
	uploadCmd.Flags().StringP("localPath", "l", "", "The path of the local server.")
	uploadCmd.Flags().BoolP("cover", "", false, "Is the remote file or directory with the same name overwritten. If not covered, create and save according to the file name and IP address. For example: info.log_192.168.200.2")
}

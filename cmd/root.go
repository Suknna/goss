/*
Copyright © 2024 Suknna

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	timeout      int
	remoteUser   string
	remotePasswd string
	address      string
	sudoPass     string
	auto         bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "goss",
	Short:   "ssh/sftp tools",
	Version: "v0.1.0 by Suknna",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle.")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 180, "Maximum waiting time for SSH.")
	rootCmd.PersistentFlags().StringVar(&remoteUser, "user", "", "Remote connection username.")
	rootCmd.PersistentFlags().StringVar(&remotePasswd, "pass", "", "Remote connection user password.")
	rootCmd.PersistentFlags().StringVar(&address, "address", "", "The address for remote connection to the server, if you want to specify a port, please add a colon and port after the address. For example: 127.0.0.1:22 .")
	rootCmd.PersistentFlags().BoolVar(&auto, "auto", false, "Whether to automatically skip remote private key verification.")
	rootCmd.PersistentFlags().StringVar(&sudoPass, "sudopass", "", "Password for privileged account.")
}

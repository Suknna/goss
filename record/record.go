package record

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

type SftpTaskRecordArgs func(args ...interface{}) error
type SSHRecordArgs func(args ...interface{}) (string, error)

func TaskLog(name string) {
	terminalWidth := getTerminalWidth()
	message := fmt.Sprintf("%s Task: [%s] ", time.Now().Format("2006-01-02 15:04:05"), name)
	paddingSize := terminalWidth - len(message)
	if paddingSize > 0 {
		padding := strings.Repeat("-", paddingSize)
		fmt.Println(message + padding)
	} else {
		fmt.Println(message)
	}
}

func SuccessPrint(ip string) {
	color.Green("%s SUCCESS", ip)
}

func SftpTaskRecord(ip, mode string, f SftpTaskRecordArgs) {
	start := time.Now()
	var err error
	defer func() {
		elapsed := time.Since(start) // 计算执行时间
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
		}
		if err != nil {
			color.Red("%s FAIL ==> %s %v [%s] failed with error: %v\n", ip, start.Format("2006-01-02 15:04:05"), elapsed, mode, err)
		} else {
			color.Green("%s PASS\n", ip)
		}
	}()
	err = f()
}

func SSHRecord(ip, mode string, f SSHRecordArgs) {
	start := time.Now()
	var (
		err error
		out string
	)
	defer func() {
		elapsed := time.Since(start) // 计算执行时间
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
		}
		if err != nil {
			color.Red("%s FAIL ==> %s %v [%s] failed with error: %v\n", ip, start.Format("2006-01-02 15:04:05"), elapsed, mode, err)
		} else {
			color.Green("%s PASS ==> \n%s\t", ip, out)
		}
	}()
	out, err = f()
}

// 获取终端宽度
func getTerminalWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80 // 默认宽度为80
	}
	parts := strings.Fields(string(out))
	width, err := strconv.Atoi(parts[1])
	if err != nil {
		return 80 // 默认宽度为80
	}
	return width
}

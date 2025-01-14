package engine

import (
	"fmt"
	"goss/auth"
	"goss/cfg"
	"goss/record"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
)

type host struct {
	ip       string
	port     string
	sudopass string
}

func AutomaticExecution(octopus *cfg.Goss) {
	// 创建row
	var (
		ru, rp string
	)
	// 创建表格
	t := table.NewWriter()
	t.SetTitle("command or script output count")
	t.SetOutputMirror(os.Stdout)
	// 创建表头
	t.AppendHeader(table.Row{
		"IP",
		"Run",
		"Out",
		"Err",
		"Status",
	})
	for _, run := range octopus.Runs {
		record.TaskLog(run.Name)
		if len(run.Groups) == 0 {
			for _, group := range octopus.Groups {
				if group.RemoteUser == "" && group.RemotePasswd == "" {
					ru = octopus.Global.RemoteUser
					rp = octopus.Global.RemotePasswd
				} else {
					ru = group.RemoteUser
					rp = group.RemotePasswd
				}

				executeTasks(run, createHosts(group), ru, rp, octopus.Global.TimeOut, octopus.Global.WorkLoad, t)
			}
		}
		for _, runGroup := range run.Groups {
			for _, group := range octopus.Groups {
				if group.Name != runGroup {
					continue
				}
				if group.RemoteUser == "" && group.RemotePasswd == "" {
					ru = octopus.Global.RemoteUser
					rp = octopus.Global.RemotePasswd
				} else {
					ru = group.RemoteUser
					rp = group.RemotePasswd
				}
				executeTasks(run, createHosts(group), ru, rp, octopus.Global.TimeOut, octopus.Global.WorkLoad, t)
			}
		}
	}
	t.Render()
}

func createHosts(group *cfg.Group) []*host {
	var hosts []*host
	if len(group.SudoPass) == 0 {
		for i := 0; i < len(group.Address); i++ {
			ip, port := cfg.ParserIP(group.Address[i])
			hosts = append(hosts, &host{
				ip:       ip,
				port:     port,
				sudopass: "",
			})
		}
	} else {
		for i := 0; i < len(group.Address); i++ {
			ip, port := cfg.ParserIP(group.Address[i])
			hosts = append(hosts, &host{
				ip:       ip,
				port:     port,
				sudopass: group.SudoPass[i],
			})
		}
	}

	return hosts
}

func executeTasks(run *cfg.Run, hosts []*host, user, pass string, timeOut int, workLoad int, t table.Writer) {
	var wg sync.WaitGroup
	// slog.Info("进入executeTaks中，开始执行任务")
	var goroutineLimiter = make(chan uint, workLoad)
	for _, h := range hosts {
		// slog.Info("进入循环")
		wg.Add(1)
		goroutineLimiter <- 0
		go func(h *host, u, p string) {
			defer func() {
				<-goroutineLimiter
				wg.Done()
			}()
			// slog.Info("进入goroutine内部，匹配运行哪个命令")
			if run.Execute != nil {
				var status string
				start := time.Now()
				// slog.Info("进入批量执行命令的操作")
				out, err := Execute(&auth.Opts{
					Port:           h.port,
					IP:             h.ip,
					RemoteUser:     u,
					RemotePassword: p,
					TimeOut:        timeOut,
				}, h.sudopass, run.Execute.Command)
				elapsed := time.Since(start) // 计算执行时间
				if err != nil {
					color.Red("%s FAIL ==> %s %v [%s] failed with error: %v\n", h.ip, start.Format("2006-01-02 15:04:05"), elapsed, "execute", err)
					status = ":("
				} else {
					color.Green("%s PASS\n", h.ip)
					status = ":)"
				}
				t.AppendRow(table.Row{
					h.ip,
					run.Execute.Command,
					out,
					err,
					status,
				})
			} else if run.Download != nil {
				var localPath = run.Download.LocalPath
				if run.Download.Cover {
					localPath = fmt.Sprintf("%s_%s", run.Download.RemotePath, h.ip)
				}
				record.SftpTaskRecord(h.ip, "download", func(args ...interface{}) error {
					return Download(&auth.Opts{
						Port:           h.port,
						IP:             h.ip,
						RemoteUser:     u,
						RemotePassword: p,
						TimeOut:        timeOut,
					}, run.Download.RemotePath, localPath)
				})
			} else if run.Upload != nil {
				var remotePath = run.Upload.RemotePath
				if run.Upload.Cover {
					remotePath = fmt.Sprintf("%s_%s", run.Upload.LocalPath, h.ip)
				}
				record.SftpTaskRecord(h.ip, "upload", func(args ...interface{}) error {
					return Upload(&auth.Opts{
						Port:           h.port,
						IP:             h.ip,
						RemoteUser:     u,
						RemotePassword: p,
						TimeOut:        timeOut,
					}, remotePath, run.Upload.LocalPath)
				})
			} else {
				slog.Info("未匹配到任何命令")
			}
		}(h, user, pass)
	}
	wg.Wait()
	// slog.Info("任务执行完成跳出executeTask函数")
	close(goroutineLimiter)
}

package dispatcher

/*
核心分发逻辑，对外暴露统一的api
管理ssh连接池
*/

import (
	"context"
	"errors"
	"fmt"
	"goss/internal/config"
	"goss/internal/model"
	"goss/internal/printer"
	"goss/internal/transfer"
	"goss/internal/utils"
	"goss/internal/xerrors"
	"goss/pkg/easyssh"
	"log/slog"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
)

func Run(hosts []*config.Host, tasks []*config.Task, cfg *config.GossConfig, save string) {
	// 初始化资源状态跟踪
	startTime := time.Now()
	totalTasks := len(hosts) * len(tasks)
	var (
		completedTasks int32 // 记录全局成功任务数量
		failedTasks    int32 // 记录全局失败任务数量
	)
	// 记录初始
	slog.Info("Task initialization started", //任务初始化开始
		"Total number of hosts", len(hosts), //总主机数
		"Number of tasks per host", len(tasks), //每主机任务数
		"Total number of tasks", totalTasks, //总任务数
		"Maximum concurrency", cfg.Execution.MaxWorkers, //最大并发数
		"Stop on error mode", cfg.Execution.StopOnError) //停止错误模式
	// 创建工作池
	maxWorkersCh := make(chan struct{}, cfg.Execution.MaxWorkers)
	// 结果收集通道
	resultCh := make(chan model.HostTask, len(hosts))
	// 创建等待组
	var wg sync.WaitGroup
	stopProgress := make(chan struct{})
	defer close(stopProgress)
	// 处理每个主机
	for i, host := range hosts {
		wg.Add(1)
		maxWorkersCh <- struct{}{}
		go func(host *config.Host, i int) {
			defer func() {
				<-maxWorkersCh
				wg.Done()
			}()
			// 为主机运行任务
			results := taskRun(host, tasks, cfg, i, &completedTasks, &failedTasks)
			// 创建结果收集结构体
			resultCh <- model.HostTask{
				Index:   i,
				HostIP:  host.IP,
				Results: results,
			}
		}(host, i)
	}
	// 当所有goroutine完成时关闭结果通道
	wg.Wait()
	close(resultCh)
	totalTime := time.Since(startTime)
	slog.Info("All tasks have been completed.", //所有任务已完成
		"Total time consumed", totalTime.Round(time.Second), //总耗时
		"Total number of tasks", totalTasks, //总任务数
		"Successful tasks", atomic.LoadInt32(&completedTasks), //成功任务
		"Failed tasks", atomic.LoadInt32(&failedTasks), //失败任务
		"Average speed", fmt.Sprintf("%.1f Tasks per second", float64(totalTasks)/totalTime.Seconds()), //平均速度 任务/秒
		"Average delay", fmt.Sprintf("%v/Task", (totalTime/time.Duration(totalTasks)).Round(time.Millisecond))) //平均延迟/任务

	slog.Info("All tasks have been completed. Collecting information and printing results now........")
	printer.PrintDivider()
	// 收集并处理结果，按ID排序
	var HostTasks []*model.HostTask
	for result := range resultCh {
		HostTasks = append(HostTasks, &result)
	}
	// 按ID排序
	sort.Slice(HostTasks, func(i, j int) bool {
		return HostTasks[i].Index < HostTasks[j].Index
	})
	// 将排序后的结果保存到json,excel文件或者打印到终端
	printer.PrintResults(HostTasks, printer.Format(save))
}

func taskRun(host *config.Host, tasks []*config.Task, cfg *config.GossConfig, goroutineID int, completedTasks, failedTasks *int32) []*model.TaskResult {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Task coroutine crashed",
				"goroutine", goroutineID,
				"host", host.IP,
				"error", r,
				"stack", string(debug.Stack()))
			atomic.AddInt32(failedTasks, int32(len(tasks)))
		}
	}()

	isConnectedSuccessfully := true
	client, createSSHErr := easyssh.NewClient(easyssh.Opts{
		IP:             host.IP,
		Port:           host.Port,
		User:           host.User,
		Passwd:         host.Password,
		ConnectTimeout: cfg.Connection.ConnectTimeout,
		Mode:           easyssh.SecurityMode(cfg.Connection.SecurityMode),
	})
	if createSSHErr != nil {
		isConnectedSuccessfully = false
		wrappedErr := xerrors.ConnectionErr(
			"ssh_connect",
			host.IP,
			createSSHErr,
		).WithDetails(map[string]interface{}{
			"port":     host.Port,
			"timeout":  cfg.Connection.ConnectTimeout,
			"attempts": 1,
		})

		slog.Error("SSH connection failed",
			"host", host.IP,
			"error", wrappedErr,
			"details", wrappedErr.Details)
	}
	var (
		canProceed bool
		results    []*model.TaskResult
	)

	for _, task := range tasks {
		taskStartTime := time.Now()
		var result *model.TaskResult

		if !isConnectedSuccessfully {
			results = append(results, &model.TaskResult{
				Task:   *task,
				StdErr: fmt.Errorf("the task cannot proceed due to the inability to establish an SSH connection. " + createSSHErr.Error()),
			})
			continue
		}
		if shouldSkipTask(cfg, canProceed) {
			result = &model.TaskResult{
				Task:   *task,
				StdErr: errors.New("the pre-task execution failed, the current task will be skipped"),
			}
			results = append(results, result)
			continue
		}
		switch task.Type {
		case config.CMD:
			result = command(client, cfg.Execution.TaskTimeout, host.SudoPass, *task)
		case config.SCRIPT:
			result = script(client, cfg.Execution.TaskTimeout, host.SudoPass, *task)
		case config.UPLOAD:
			result = upload(client, cfg.FileTransfer.TransferTimeout, *task, cfg.FileTransfer.Retries, cfg.FileTransfer.OverwritePolicy)
		case config.DOWNLOAD:
			result = download(client, cfg.FileTransfer.TransferTimeout, *task, cfg.FileTransfer.Retries, cfg.FileTransfer.OverwritePolicy)
		}
		// 输出任务结果
		if result.StdErr != nil {
			if gerr, ok := result.StdErr.(*xerrors.GossError); ok {
				slog.Error("Task failed",
					"Worker", goroutineID,
					"Host", host.IP,
					"Task", task.Description,
					"ErrorType", gerr.Type,
					"Error", gerr.Error(),
					"Details", gerr.Details,
					"Time-consuming", time.Since(taskStartTime).Round(time.Millisecond))
			} else {
				slog.Error("Task failed",
					"Worker", goroutineID,
					"Host", host.IP,
					"Task", task.Description,
					"Error", result.StdErr,
					"Time-consuming", time.Since(taskStartTime).Round(time.Millisecond))
			}
			atomic.AddInt32(failedTasks, 1)
		} else {
			slog.Info("Task completed",
				"Worker", goroutineID,
				"Host", host.IP,
				"Task", task.Description,
				"Time-consuming", time.Since(taskStartTime).Round(time.Millisecond))
			atomic.AddInt32(completedTasks, 1)
		}
		// 收集所有结果
		results = append(results, result)
	}
	return results
}

func command(client *ssh.Client, timeout int, passwd string, task config.Task) *model.TaskResult {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()
	// 创建session
	exe, err := easyssh.NewCTXSession(ctx, client)
	if err != nil {
		return &model.TaskResult{
			Task: task,
			StdErr: xerrors.Wrap(err, xerrors.ExecutionError,
				"create_session",
				task.Description,
				"failed to create SSH session"),
		}
	}
	var out string
	var execErr error

	if task.RequireSudo {
		out, execErr = exe.ExecutePrivilegedCommandOverSSH(task.Cmd, passwd)
	} else {
		out, execErr = exe.Execute(task.Cmd)
	}

	if execErr != nil {
		return &model.TaskResult{
			Task: task,
			StdErr: xerrors.Wrap(execErr, xerrors.ExecutionError,
				"execute_command",
				task.Cmd,
				"command execution failed"),
		}
	}

	return &model.TaskResult{
		Task:   task,
		StdOut: out,
	}
}

func script(client *ssh.Client, timeout int, passwd string, task config.Task) *model.TaskResult {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()
	// 传输文件到目标服务器
	t, err := transfer.NewTransferHandler(task.Remote, task.Local, config.Always, client)
	if err != nil {
		return &model.TaskResult{
			Task:   task,
			StdErr: fmt.Errorf("script transfer failed. %s", err.Error()),
			StdOut: "",
		}
	}
	// 传输脚本
	_, err = t.Upload(ctx)
	if err != nil {
		return &model.TaskResult{
			StdErr: fmt.Errorf("script transfer failed. %s", err.Error()),
			StdOut: "",
		}
	}
	// 执行命令
	exe, err := easyssh.NewCTXSession(ctx, client)
	if err != nil {
		return &model.TaskResult{
			Task:   task,
			StdErr: err,
			StdOut: "",
		}
	}
	if task.RequireSudo {
		out, err := exe.ExecutePrivilegedCommandOverSSH(task.Cmd, passwd)
		return &model.TaskResult{
			Task:   task,
			StdErr: err,
			StdOut: out,
		}
	} else {
		out, err := exe.Execute(task.Cmd)
		return &model.TaskResult{
			Task:   task,
			StdErr: err,
			StdOut: out,
		}
	}
}

func upload(client *ssh.Client, timeout int, task config.Task, retry int, transferPolicy config.FileTransferPolicy) *model.TaskResult {
	var (
		result model.TaskResult
		local  string
		remote string
		err    error
	)
	if utils.ContainsTemplate(result.Local) {
		local, err = utils.RenderPathTemplate(task.Local, strings.Split(client.LocalAddr().String(), ":")[0])
		if err != nil {
			result = model.TaskResult{
				Task:   task,
				StdErr: err,
			}
			return &result
		}
	}
	if utils.ContainsTemplate(result.Remote) {
		remote, err = utils.RenderPathTemplate(task.Remote, strings.Split(client.LocalAddr().String(), ":")[0])
		if err != nil {
			result = model.TaskResult{
				Task:   task,
				StdErr: err,
			}
			return &result
		}
	}
	t, err := transfer.NewTransferHandler(remote, local, transferPolicy, client)
	if err != nil {
		return &model.TaskResult{
			Task: task,
			StdErr: xerrors.Wrap(err, xerrors.ResourceError,
				"init_transfer",
				fmt.Sprintf("%s->%s", local, remote),
				"failed to initialize file transfer"),
		}
	}
	var lastErr error
	for i := 1; i <= retry; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
		defer cancel()

		_, err := t.Upload(ctx)
		if err == nil {
			return &model.TaskResult{
				Task:   task,
				StdOut: "Upload successful",
			}
		}

		lastErr = xerrors.Wrap(err, xerrors.ResourceError,
			"file_upload",
			remote,
			fmt.Sprintf("upload failed (attempt %d/%d)", i, retry))
	}

	return &model.TaskResult{
		Task:   task,
		StdErr: lastErr,
	}
}

func download(client *ssh.Client, timeout int, task config.Task, retry int, transferPolicy config.FileTransferPolicy) *model.TaskResult {
	var (
		result model.TaskResult
		local  string
		remote string
		err    error
	)
	if utils.ContainsTemplate(result.Local) {
		local, err = utils.RenderPathTemplate(task.Local, strings.Split(client.LocalAddr().String(), ":")[0])
		if err != nil {
			result = model.TaskResult{
				Task:   task,
				StdErr: err,
			}
			return &result
		}
	}
	if utils.ContainsTemplate(result.Remote) {
		remote, err = utils.RenderPathTemplate(task.Remote, strings.Split(client.LocalAddr().String(), ":")[0])
		if err != nil {
			result = model.TaskResult{
				Task:   task,
				StdErr: err,
			}
			return &result
		}
	}
	t, err := transfer.NewTransferHandler(remote, local, transferPolicy, client)
	if err != nil {
		return &model.TaskResult{
			Task: task,
			StdErr: xerrors.Wrap(err, xerrors.ResourceError,
				"init_transfer",
				fmt.Sprintf("%s->%s", remote, local),
				"failed to initialize file transfer"),
		}
	}
	var lastErr error
	for i := 1; i <= retry; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
		defer cancel()

		_, err := t.Download(ctx)
		if err == nil {
			return &model.TaskResult{
				Task:   task,
				StdOut: "Download successful",
			}
		}

		lastErr = xerrors.Wrap(err, xerrors.ResourceError,
			"file_download",
			remote,
			fmt.Sprintf("download failed (attempt %d/%d)", i, retry))
	}

	return &model.TaskResult{
		Task:   task,
		StdErr: lastErr,
	}
}

func shouldSkipTask(cfg *config.GossConfig, canProceed bool) bool {
	if cfg.Execution.StopOnError && canProceed {
		return true
	}
	return false
}

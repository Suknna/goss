package model

import "goss/internal/config"

type HostTask struct {
	Index   int           // 主机的索引，由于主机执行顺序是并发执行通过这个index在输出时进行排序
	HostIP  string        // 记录主机的ip信息
	Results []*TaskResult // 记录主机任务执行情况
}

type TaskResult struct {
	config.Task        // 继承于config包的任务配置
	StdErr      error  // 当前任务执行失败原因
	StdOut      string // 当前任务执行成功的信息
}

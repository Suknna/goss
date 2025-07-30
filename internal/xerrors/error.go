package xerrors

import (
	"strings"
	"time"
)

type ErrorType string

// 错误类型
const (
	ConnectionError    ErrorType = "connection"    // 连接错误
	ExecutionError     ErrorType = "execution"     // 执行错误
	TimeoutError       ErrorType = "timeout"       // 任务超时
	PermissionError    ErrorType = "permission"    // 权限错误
	ValidationError    ErrorType = "validation"    // 验证错误
	ResourceError      ErrorType = "resource"      // 重试多次失败
	ConfigurationError ErrorType = "configuration" // 配置错误
)

// 基础错误结构体
type GossError struct {
	Type      ErrorType              // 错误类型
	Operation string                 // 操作类型
	Target    string                 // 操作目标
	Message   string                 // 错误信息
	Timestamp time.Time              // 发生时间
	Cause     error                  // 原始错误
	Details   map[string]interface{} // 额外上下文
}

// error接口
func (e *GossError) Error() string {
	var sb strings.Builder
	sb.WriteString(string(e.Type)) // 转换类型
	sb.WriteString(" error")
	if e.Operation != "" {
		sb.WriteString(" in ")
		sb.WriteString(e.Operation) // 写入操作类型
	}
	if e.Target != "" {
		sb.WriteString(" on ")
		sb.WriteString(e.Target)
	}
	sb.WriteString(": ")
	sb.WriteString(e.Message)
	if e.Cause != nil {
		sb.WriteString(" (cause: ")
		sb.WriteString(e.Cause.Error())
		sb.WriteString(")")
	}
	return sb.String()
}

// 解包获取原始错误
func (e *GossError) Unwrap() error {
	return e.Cause
}

// 构造函数
func New(errType ErrorType, op, target, msg string) *GossError {
	return &GossError{
		Type:      errType,
		Operation: op,
		Target:    target,
		Message:   msg,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}
}

// 包装现有错误
func Wrap(err error, errType ErrorType, op, target, msg string) *GossError {
	return &GossError{
		Type:      errType,
		Operation: op,
		Target:    target,
		Message:   msg,
		Timestamp: time.Now(),
		Cause:     err,
		Details:   make(map[string]interface{}),
	}
}

// 快捷构造函数
func ConnectionErr(op, target string, cause error) *GossError {
	return Wrap(cause, ConnectionError, op, target, "connection failed")
}

// 执行命令错误
func ExecutionErr(op, target string, cause error) *GossError {
	return Wrap(cause, ExecutionError, op, target, "execution failed")
}

// 超时
func TimeoutErr(op, target string) *GossError {
	return New(TimeoutError, op, target, "operation timed out")
}

// 添加上下文详情
func (e *GossError) WithDetails(details map[string]interface{}) *GossError {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// 检查错误类型
func IsType(err error, t ErrorType) bool {
	if e, ok := err.(*GossError); ok {
		return e.Type == t
	}
	return false
}

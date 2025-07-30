package utils

import (
	"bytes"
	"strings"
	"text/template"
	"time"
)

// TemplateData 模板变量结构体
type TemplateData struct {
	IP   string // 主机IP地址
	TIME string // 执行时间戳(格式: YYYYMMDD_HHmmss)
}

// RenderPathTemplate 路径模板渲染
func RenderPathTemplate(tpl string, ip string) (string, error) {
	data := TemplateData{
		IP:   ip,
		TIME: time.Now().Format("20060102_150405"),
	}

	tmpl, err := template.New("path").Parse(tpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ContainsTemplate 检查字符串是否包含模板语法
func ContainsTemplate(s string) bool {
	return strings.Contains(s, "{{") && strings.Contains(s, "}}")
}

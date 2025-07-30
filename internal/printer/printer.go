package printer

import (
	"encoding/json"
	"fmt"
	"goss/internal/model"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/xuri/excelize/v2"
	"golang.org/x/sys/unix"
)

type Format string

const (
	FormatJSON  Format = "json"
	FormatExcel Format = "excel"
	FormatTable Format = "table"
)

func PrintResults(hostResults []*model.HostTask, format Format) {
	switch format {
	case FormatJSON:
		printJSON(hostResults)
	case FormatExcel:
		printExcel(hostResults)
	default:
		printTable(hostResults)
	}
}

func printJSON(results []*model.HostTask) {
	slog.Info("Starting to save JSON information...")
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Export json failed, err: %d", err)
		return
	}
	fileName := time.Now().Format("2006-01-02T150405") + ".json"
	fileP := path.Join(dir, fileName)
	f, err := os.OpenFile(fileP, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		slog.Error("Failed to save the output content as a JSON file.", slog.String("ERROR", err.Error()))
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()
	if err := json.NewEncoder(f).Encode(results); err != nil {
		slog.Error("Failed to encode into JSON information.", slog.String("ERROR", err.Error()))
	} else {
		slog.Info("The JSON file has been successfully generated.", slog.String("PATH", fileP))
	}
}

// 打印表格
func printTable(hosts []*model.HostTask) {
	slog.Info("Starting to print the table...")
	headers := table.Row{
		"ID",
		"Type",
		"Describe",
		"State",
		"Output",
		"Error",
	}
	for _, ht := range hosts {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		fmt.Printf("\nhost: %s (index: %d) \n", ht.HostIP, ht.Index)
		t.AppendHeader(headers)
		for i, result := range ht.Results {
			status := ":)"
			detail := firstLine(result.StdOut)
			if result.StdErr != nil {
				status = ":("
			}
			t.AppendRow(
				table.Row{
					i + 1,
					result.Type,
					result.Description,
					status,
					detail,
					result.StdErr,
				},
				table.RowConfig{},
			)
		}
		// 结束表格
		t.Render()
	}
	slog.Info("The table printing is completed.")
}
func firstLine(s string) string {
	if idx := strings.Index(s, "\n"); idx != -1 {
		return s[:idx]
	}
	return s
}

func printExcel(hostsResults []*model.HostTask) {
	slog.Info("Starting to generate Excel document...")
	// 获取当前绝对路径
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Export excel failed, err: %d", err)
		return
	}
	// 获取当前时间来创建excel文件名称,精确到ms
	fileName := time.Now().Format("2006-01-02T150405") + ".xlsx"
	// 拼接文件名称
	fileP := path.Join(dir, fileName)
	f := excelize.NewFile()
	// 表头
	headers := []string{"ID", "Host", "Task ID", "Task Description", "Task Type", "Task Info", "Output", "State"}
	sheet := "Sheet1" // sheet名称
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheet, cell, header)
	}
	// 写入数据并记录合并区域
	row := 2 // 第二行
	for _, hostResult := range hostsResults {
		startRow := row
		// 任务数量
		taskCount := len(hostResult.Results)
		for id, result := range hostResult.Results {
			var (
				info   string
				state  = ":)"
				output string
			)
			if result.Type == "upload" || result.Type == "download" {
				info = fmt.Sprintf("%s: local=%s, remote=%s", result.Type, result.Local, result.Remote)
			} else {
				info = result.Cmd
			}

			if result.StdErr != nil {
				state = ":("
				if result.StdOut == "" {
					output = result.StdErr.Error()
				} else {
					output = fmt.Sprintf("OUT:\n%s\n\nERR:\n%s", strings.TrimSpace(result.StdOut), strings.TrimSpace(result.StdErr.Error()))
				}
			} else {
				output = strings.TrimSpace(result.StdOut)
			}
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), hostResult.Index)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), hostResult.HostIP)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), id+1)
			f.SetCellValue(sheet, fmt.Sprintf("D%d", row), result.Description)
			f.SetCellValue(sheet, fmt.Sprintf("E%d", row), result.Type)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", row), info)
			f.SetCellValue(sheet, fmt.Sprintf("G%d", row), output)
			f.SetCellValue(sheet, fmt.Sprintf("H%d", row), state)
			row++
		}
		if taskCount > 1 {
			endRow := startRow + taskCount - 1
			f.MergeCell(sheet, fmt.Sprintf("A%d", startRow), fmt.Sprintf("A%d", endRow))
			f.MergeCell(sheet, fmt.Sprintf("B%d", startRow), fmt.Sprintf("B%d", endRow))
		}
	}
	// 设置单元格样式
	style, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "top",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
	})
	if err != nil {
		slog.Warn("Failed to generate global table style", slog.String("Tips", err.Error()))
	}
	styleHeader, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"3366CC"},
		},
	})
	if err != nil {
		slog.Warn("Failed to create header table style", slog.String("Tips", err.Error()))
	}
	err = f.SetCellStyle(sheet, "A1", fmt.Sprintf("H%d", row-1), style)
	if err != nil {
		slog.Warn("Failed to set overall table style", slog.String("Tips", err.Error()))
	}
	err = f.SetCellStyle(sheet, "A1", "H1", styleHeader)
	if err != nil {
		slog.Warn("Failed to set header table style", slog.String("Tips", err.Error()))
	}
	// 写入文件
	if err := f.SaveAs(fileP); err != nil {
		slog.Error("Excel document generation failed.", slog.String("ERROR", err.Error()))
	} else {
		slog.Info("The Excel document has been successfully generated.", slog.String("PATH", fileP))
	}
}

// 打印分割线
func PrintDivider() {
	dividerLen := 80
	ws, _ := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	dividerLen = int(ws.Col)
	fmt.Printf("\n%s\n", strings.Repeat("-", dividerLen))
}

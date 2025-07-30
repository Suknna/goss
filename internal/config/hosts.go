package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Host struct {
	IP       string
	Port     string
	User     string
	Password string
	SudoPass string
}

func ParseHostsFile(path string) ([]*Host, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hosts []*Host
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 分割字段
		parts := strings.Split(line, ",")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid format at line %d", lineNum)
		}
		port := strconv.Itoa(DefaultPort)
		if strings.Contains(parts[0], ":") {
			port = strings.Split(parts[0], ":")[1]
		}
		host := Host{
			IP:       strings.TrimSpace(parts[0]),
			Port:     port,
			User:     strings.TrimSpace(parts[1]),
			Password: strings.TrimSpace(parts[2]),
			SudoPass: strings.TrimSpace(parts[3]),
		}
		hosts = append(hosts, &host)
	}
	return hosts, scanner.Err()
}

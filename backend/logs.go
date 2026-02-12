package main

import (
	"bufio"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const (
	defaultLogLines = 100
	maxLogLines     = 2000
)

// HandleBackendLogs 返回后端日志尾部内容。
func HandleBackendLogs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		lines := parseLinesParam(r.URL.Query().Get("lines"))
		logPath, err := resolveBackendLogPath()
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		items, err := readTailLines(logPath, lines)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "read log failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"lines": len(items),
			"text":  joinLines(items),
		})
	}
}

// parseLinesParam 解析日志行数参数。
func parseLinesParam(raw string) int {
	if raw == "" {
		return defaultLogLines
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return defaultLogLines
	}
	if value > maxLogLines {
		return maxLogLines
	}
	return value
}

// resolveBackendLogPath 定位后端日志文件路径。
func resolveBackendLogPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", errors.New("cannot resolve log path")
	}
	candidates := []string{
		filepath.Join(wd, "backend.out"),
		// 重要逻辑：兼容从仓库根目录启动时日志落在 backend/backend.out 的情况。
		filepath.Join(wd, "backend", "backend.out"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", errors.New("log file not found")
}

// readTailLines 读取文件最后 N 行。
func readTailLines(path string, lines int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 重要逻辑：使用环形缓冲，仅保留最后 N 行，避免读取超大文件占用过多内存。
	buffer := make([]string, lines)
	count := 0
	idx := 0
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		buffer[idx] = scanner.Text()
		idx = (idx + 1) % lines
		count++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if count == 0 {
		return []string{}, nil
	}
	if count < lines {
		return buffer[:count], nil
	}
	result := make([]string, 0, lines)
	result = append(result, buffer[idx:]...)
	result = append(result, buffer[:idx]...)
	return result, nil
}

// joinLines 将日志行拼接为文本。
func joinLines(items []string) string {
	if len(items) == 0 {
		return ""
	}
	out := make([]byte, 0, len(items)*64)
	for i, line := range items {
		out = append(out, line...)
		if i < len(items)-1 {
			out = append(out, '\n')
		}
	}
	return string(out)
}

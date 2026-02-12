package main

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// HandleGitStatus 返回当前目录的 Git 状态。
func HandleGitStatus(manager *SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			writeError(w, http.StatusBadRequest, "session_id required")
			return
		}
		sessionCWD, err := resolveSessionCWD(manager, sessionID)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		root, err := resolveGitRoot(sessionCWD)
		if err != nil {
			writeError(w, http.StatusNotFound, "git status failed")
			return
		}
		cmd := exec.Command("git", "-C", root, "status", "--porcelain=v1", "-z")
		output, err := cmd.Output()
		if err != nil {
			writeError(w, http.StatusNotFound, "git status failed")
			return
		}
		items := parseGitStatus(output)
		for _, item := range items {
			path, ok := item["path"].(string)
			if !ok || path == "" {
				continue
			}
			additions, deletions := gitDiffStatForPath(root, path)
			item["additions"] = additions
			item["deletions"] = deletions
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

// HandleGitDiff 返回指定文件的 diff。
func HandleGitDiff(manager *SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			writeError(w, http.StatusBadRequest, "session_id required")
			return
		}
		path := r.URL.Query().Get("path")
		if path == "" {
			writeError(w, http.StatusBadRequest, "path required")
			return
		}
		sessionCWD, err := resolveSessionCWD(manager, sessionID)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		root, err := resolveGitRoot(sessionCWD)
		if err != nil {
			writeError(w, http.StatusNotFound, "git diff failed")
			return
		}
		diff, err := gitDiffForPath(root, path)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "git diff failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"diff": diff})
	}
}

// parseGitStatus 解析 git status -z 输出。
func parseGitStatus(output []byte) []map[string]any {
	entries := bytes.Split(output, []byte{0})
	items := make([]map[string]any, 0)
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if len(entry) == 0 {
			continue
		}
		if len(entry) < 4 {
			continue
		}
		status := string(entry[:2])
		path := string(entry[3:])
		item := map[string]any{
			"path":   path,
			"status": strings.TrimSpace(status),
		}
		// 重要逻辑：处理重命名/复制的第二个路径。
		if strings.Contains(status, "R") || strings.Contains(status, "C") {
			if i+1 < len(entries) && len(entries[i+1]) > 0 {
				item["orig_path"] = path
				item["path"] = string(entries[i+1])
				i++
			}
		}
		items = append(items, item)
	}
	return items
}

// gitDiffForPath 获取指定路径的 diff 文本。
func gitDiffForPath(root, path string) (string, error) {
	normalized := normalizeGitPath(root, path)
	if isTrackedGitFile(root, normalized) {
		output, err := runGitDiff("git", "-C", root, "diff", "--", normalized)
		return output, err
	}
	// 重要逻辑：未跟踪文件使用 no-index diff。
	output, err := runGitDiff("git", "-C", root, "diff", "--no-index", "--", "/dev/null", normalized)
	return output, err
}

// gitDiffStatForPath 获取指定路径的增删行统计。
func gitDiffStatForPath(root, path string) (int, int) {
	normalized := normalizeGitPath(root, path)
	var output string
	var err error
	if isTrackedGitFile(root, normalized) {
		output, err = runGitDiff("git", "-C", root, "diff", "--numstat", "--", normalized)
	} else {
		// 重要逻辑：未跟踪文件使用 no-index 统计，保证新增文件也能返回行数。
		output, err = runGitDiff("git", "-C", root, "diff", "--numstat", "--no-index", "--", "/dev/null", normalized)
	}
	if err != nil {
		return 0, 0
	}
	return parseNumStat(output)
}

// parseNumStat 解析 numstat 的增删行数据。
func parseNumStat(output string) (int, int) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		additions := parseNumStatValue(parts[0])
		deletions := parseNumStatValue(parts[1])
		return additions, deletions
	}
	return 0, 0
}

// parseNumStatValue 解析 numstat 的数值字段，处理 "-" 或非法情况。
func parseNumStatValue(value string) int {
	if value == "-" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

// isTrackedGitFile 判断文件是否在 Git 索引中。
func isTrackedGitFile(root, path string) bool {
	cmd := exec.Command("git", "-C", root, "ls-files", "--error-unmatch", "--", path)
	return cmd.Run() == nil
}

// normalizeGitPath 规范化 git diff 使用的路径，避免重复前缀导致找不到文件。
func normalizeGitPath(root, path string) string {
	clean := filepath.Clean(filepath.FromSlash(path))
	target := filepath.Join(root, clean)
	if _, err := os.Stat(target); err == nil {
		return filepath.ToSlash(clean)
	}
	base := filepath.Base(root)
	prefix := base + string(filepath.Separator)
	if strings.HasPrefix(clean, prefix) {
		trimmed := strings.TrimPrefix(clean, prefix)
		altTarget := filepath.Join(root, trimmed)
		if _, err := os.Stat(altTarget); err == nil {
			return filepath.ToSlash(trimmed)
		}
	}
	return filepath.ToSlash(clean)
}

// resolveGitRoot 获取指定目录所在的 Git 仓库根目录。
func resolveGitRoot(cwd string) (string, error) {
	cmd := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// runGitDiff 执行 git diff 并容忍差异导致的退出码。
func runGitDiff(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		// 重要逻辑：git diff 有差异时返回 1，但输出依然有效。
		return string(output), nil
	}
	return "", err
}

package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxUploadSize = 50 * 1024 * 1024
const maxPreviewSize = 1 * 1024 * 1024

// HandleFileTree 返回指定目录下的文件列表。
func HandleFileTree(manager *SessionManager) http.HandlerFunc {
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
		root, err := resolveSessionCWD(manager, sessionID)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		relPath := r.URL.Query().Get("path")
		target, err := resolveSafePath(root, relPath)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		entries, err := os.ReadDir(target)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "read dir failed")
			return
		}
		items := make([]map[string]any, 0, len(entries))
		for _, entry := range entries {
			info, infoErr := entry.Info()
			if infoErr != nil {
				continue
			}
			items = append(items, map[string]any{
				"name":   entry.Name(),
				"path":   filepath.ToSlash(filepath.Join(relPath, entry.Name())),
				"is_dir": entry.IsDir(),
				"size":   info.Size(),
				"mtime":  info.ModTime().Unix(),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"root":    root,
			"entries": items,
		})
	}
}

// HandleFileUpload 处理文件上传。
func HandleFileUpload(manager *SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			writeError(w, http.StatusBadRequest, "session_id required")
			return
		}
		root, err := resolveSessionCWD(manager, sessionID)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		relPath := r.URL.Query().Get("path")
		targetDir, err := resolveSafePath(root, relPath)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		// 重要逻辑：限制上传大小，避免占满磁盘。
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			writeError(w, http.StatusBadRequest, "invalid multipart data")
			return
		}
		form := r.MultipartForm
		if form == nil || len(form.File) == 0 {
			writeError(w, http.StatusBadRequest, "file required")
			return
		}
		for _, headers := range form.File {
			for _, header := range headers {
				if header == nil {
					continue
				}
				src, err := header.Open()
				if err != nil {
					writeError(w, http.StatusBadRequest, "open file failed")
					return
				}
				defer src.Close()
				filename := filepath.Base(header.Filename)
				destPath := filepath.Join(targetDir, filename)
				dst, err := os.Create(destPath)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "save file failed")
					return
				}
				if _, err := io.Copy(dst, src); err != nil {
					_ = dst.Close()
					writeError(w, http.StatusInternalServerError, "save file failed")
					return
				}
				if err := dst.Close(); err != nil {
					writeError(w, http.StatusInternalServerError, "save file failed")
					return
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

// HandleFileDownload 处理文件下载。
func HandleFileDownload(manager *SessionManager) http.HandlerFunc {
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
		root, err := resolveSessionCWD(manager, sessionID)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		relPath := r.URL.Query().Get("path")
		target, err := resolveSafePath(root, relPath)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		info, err := os.Stat(target)
		if err != nil || info.IsDir() {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", info.Name()))
		http.ServeFile(w, r, target)
	}
}

// HandleFileRead 读取小文件内容用于预览。
func HandleFileRead(manager *SessionManager) http.HandlerFunc {
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
		root, err := resolveSessionCWD(manager, sessionID)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		relPath := r.URL.Query().Get("path")
		target, err := resolveSafePath(root, relPath)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		info, err := os.Stat(target)
		if err != nil || info.IsDir() {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}
		if info.Size() > maxPreviewSize {
			writeError(w, http.StatusRequestEntityTooLarge, "file too large")
			return
		}
		data, err := os.ReadFile(target)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "read file failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"path":  relPath,
			"size":  info.Size(),
			"count": len(data),
			"text":  string(data),
		})
	}
}

// resolveSessionCWD 获取会话对应的工作目录。
func resolveSessionCWD(manager *SessionManager, sessionID string) (string, error) {
	session, ok := manager.GetSession(sessionID)
	if !ok || session == nil || session.Cmd == nil || session.Cmd.Process == nil {
		return "", errors.New("session not found")
	}
	pid := session.Cmd.Process.Pid
	if pid == 0 {
		return "", errors.New("session not found")
	}
	// 重要逻辑：通过 /proc 获取进程当前目录。
	cwd, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))
	if err != nil {
		return "", errors.New("cannot resolve session cwd")
	}
	return cwd, nil
}

// resolveSafePath 将相对路径安全拼接到根目录。
func resolveSafePath(root, rel string) (string, error) {
	if filepath.IsAbs(rel) {
		return "", errors.New("absolute path not allowed")
	}
	clean := filepath.Clean(rel)
	if clean == "." {
		return root, nil
	}
	if strings.HasPrefix(clean, "..") {
		return "", errors.New("invalid path")
	}
	target := filepath.Join(root, clean)
	return target, nil
}

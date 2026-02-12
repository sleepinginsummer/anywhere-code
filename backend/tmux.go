package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

// TmuxManager 封装对 tmux 的调用。
type TmuxManager struct {
	Path  string
	Shell string
}

// NewTmuxManager 创建 TmuxManager。
func NewTmuxManager(path, shell string) *TmuxManager {
	return &TmuxManager{
		Path:  path,
		Shell: shell,
	}
}

// CreateSession 创建新的 tmux 会话。
func (t *TmuxManager) CreateSession(sessionID string) error {
	cmd := exec.Command(t.Path, "new-session", "-d", "-s", sessionID, t.Shell)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("create tmux session failed: %w: %s", err, stderr.String())
	}

	if err := t.DisableMouse(sessionID); err != nil {
		return err
	}

	return nil
}

// HasSession 检查 tmux 会话是否存在。
func (t *TmuxManager) HasSession(sessionID string) bool {
	cmd := exec.Command(t.Path, "has-session", "-t", sessionID)
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

// KillSession 关闭 tmux 会话。
func (t *TmuxManager) KillSession(sessionID string) error {
	cmd := exec.Command(t.Path, "kill-session", "-t", sessionID)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kill tmux session failed: %w: %s", err, stderr.String())
	}

	return nil
}

// DisableMouse 关闭 tmux 的鼠标模式，避免滚轮被终端应用拦截。
func (t *TmuxManager) DisableMouse(sessionID string) error {
	cmd := exec.Command(t.Path, "set", "-t", sessionID, "mouse", "off")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("disable tmux mouse failed: %w: %s", err, stderr.String())
	}

	return nil
}

// AttachCommand 返回用于 attach 的命令。
func (t *TmuxManager) AttachCommand(sessionID string) *exec.Cmd {
	// 重要逻辑：使用 -d 强制踢掉其他已连接的 tmux client，保证可重连。
	return exec.Command(t.Path, "attach", "-d", "-t", sessionID)
}

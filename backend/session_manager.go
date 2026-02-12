package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
)

// Session 保存终端会话状态。
type Session struct {
	ID           string
	Name         string
	DisplayIndex int
	Cmd          *exec.Cmd
	PTY          *os.File
	Buffer       *RingBuffer
	LastActive   time.Time
	mu           sync.Mutex
}

// SessionManager 管理所有会话。
type SessionManager struct {
	shell            string
	bufferSize       int
	mu               sync.RWMutex
	sessions         map[string]*Session
	nextDisplayIndex int
	nameDate         string
	nameSeq          int
}

// NewSessionManager 创建 SessionManager。
func NewSessionManager(shell string, bufferSize int) *SessionManager {
	return &SessionManager{
		shell:      shell,
		bufferSize: bufferSize,
		sessions:   make(map[string]*Session),
	}
}

// CreateSession 创建新的 PTY 会话。
func (m *SessionManager) CreateSession(id string) (*Session, error) {
	cmd := exec.Command(m.shell)
	// 重要逻辑：注入颜色相关环境，避免 NO_COLOR 导致的颜色禁用。
	cmd.Env = buildColorEnv()
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	// 重要逻辑：初始化一个合理的终端尺寸，避免列数为 1 导致逐字换行。
	_ = pty.Setsize(ptmx, &pty.Winsize{Cols: 120, Rows: 30})
	m.mu.Lock()
	// 重要逻辑：确保名称计数器在同一个锁内更新，避免并发重复。
	name := m.nextSessionNameLocked()
	// 重要逻辑：仅在会话全部清空后才重置编号，避免删除后编号前移。
	if len(m.sessions) == 0 {
		m.nextDisplayIndex = 1
	}
	session := &Session{
		ID:           id,
		Name:         name,
		Cmd:          cmd,
		PTY:          ptmx,
		Buffer:       NewRingBuffer(m.bufferSize),
		LastActive:   time.Now(),
		DisplayIndex: m.nextDisplayIndex,
	}
	m.nextDisplayIndex++
	m.sessions[id] = session
	m.mu.Unlock()

	return session, nil
}

// nextSessionNameLocked 生成当天的会话名称（需要在写锁内调用）。
func (m *SessionManager) nextSessionNameLocked() string {
	// 重要逻辑：以本地日期为单位递增序号，跨天重置。
	today := time.Now().Format("2006-01-02")
	if m.nameDate != today {
		m.nameDate = today
		m.nameSeq = 0
	}
	m.nameSeq++
	return fmt.Sprintf("%s-%03d", m.nameDate, m.nameSeq)
}

// buildColorEnv 构造用于 PTY 的环境变量，强制启用颜色输出。
func buildColorEnv() []string {
	// 重要逻辑：过滤 NO_COLOR/CLICOLOR=0，避免工具主动禁用颜色。
	ignoredKeys := map[string]bool{
		"NO_COLOR": true,
	}
	env := make([]string, 0, len(os.Environ())+6)
	for _, pair := range os.Environ() {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 0 {
			continue
		}
		key := parts[0]
		if ignoredKeys[key] {
			continue
		}
		if key == "CLICOLOR" && len(parts) == 2 && parts[1] == "0" {
			continue
		}
		env = append(env, pair)
	}

	env = append(env,
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"CLICOLOR=1",
		"CLICOLOR_FORCE=1",
		"FORCE_COLOR=1",
	)
	return env
}

// GetSession 返回会话。
func (m *SessionManager) GetSession(id string) (*Session, bool) {
	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()
	return session, ok
}

// ListSessions 返回所有会话的快照信息。
func (m *SessionManager) ListSessions() []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]SessionInfo, 0, len(m.sessions))
	for _, session := range m.sessions {
		session.mu.Lock()
		info := SessionInfo{
			ID:           session.ID,
			Name:         session.Name,
			DisplayIndex: session.DisplayIndex,
			LastActive:   session.LastActive,
		}
		session.mu.Unlock()
		result = append(result, info)
	}

	// 重要逻辑：按固定编号排序，避免 map 遍历导致列表顺序跳变。
	sort.Slice(result, func(i, j int) bool {
		return result[i].DisplayIndex < result[j].DisplayIndex
	})

	return result
}

// CloseSession 关闭并移除会话。
func (m *SessionManager) CloseSession(id string) error {
	m.mu.Lock()
	session, ok := m.sessions[id]
	if ok {
		delete(m.sessions, id)
	}
	m.mu.Unlock()

	if !ok {
		return errors.New("session not found")
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.PTY != nil {
		_ = session.PTY.Close()
	}
	if session.Cmd != nil && session.Cmd.Process != nil {
		_ = session.Cmd.Process.Kill()
	}

	return nil
}

// RenameSession 更新会话名称。
func (m *SessionManager) RenameSession(id, name string) error {
	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return errors.New("session not found")
	}

	session.mu.Lock()
	session.Name = name
	session.mu.Unlock()

	return nil
}

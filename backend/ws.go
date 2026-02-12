package main

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

// WSMessage 是 WebSocket 消息结构。
type WSMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

// WebSocketHandler 处理终端连接。
func WebSocketHandler(manager *SessionManager) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			writeError(w, http.StatusBadRequest, "session_id required")
			return
		}
		session, ok := manager.GetSession(sessionID)
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		if err := handleSessionWS(r.Context(), conn, session); err != nil {
			_ = conn.WriteJSON(WSMessage{Type: "exit", Data: err.Error()})
		}
	}
}

// handleSessionWS 负责转发 WebSocket 与 PTY 会话数据。
func handleSessionWS(ctx context.Context, conn *websocket.Conn, session *Session) error {
	session.mu.Lock()
	session.LastActive = time.Now()
	ptmx := session.PTY
	session.mu.Unlock()

	outputErr := make(chan error, 1)
	inputErr := make(chan error, 1)

	// 重连时先回放缓存内容。
	if cached := session.Buffer.Snapshot(); len(cached) > 0 {
		if err := conn.WriteJSON(WSMessage{Type: "output", Data: string(cached)}); err != nil {
			return err
		}
	}

	// 读取 PTY 输出并推送给前端。
	go func() {
		buffer := make([]byte, 4096)
		for {
			n, readErr := ptmx.Read(buffer)
			if n > 0 {
				session.Buffer.Write(buffer[:n])
				if writeErr := conn.WriteJSON(WSMessage{Type: "output", Data: string(buffer[:n])}); writeErr != nil {
					outputErr <- writeErr
					return
				}
			}
			if readErr != nil {
				if readErr == io.EOF {
					outputErr <- nil
				} else {
					outputErr <- readErr
				}
				return
			}
		}
	}()

	// 读取 WebSocket 输入并写入 PTY。
	go func() {
		for {
			var msg WSMessage
			if err := conn.ReadJSON(&msg); err != nil {
				inputErr <- err
				return
			}

			session.mu.Lock()
			session.LastActive = time.Now()
			session.mu.Unlock()

			switch msg.Type {
			case "input":
				if msg.Data == "" {
					continue
				}
				if _, err := ptmx.Write([]byte(msg.Data)); err != nil {
					inputErr <- err
					return
				}
			case "resize":
				if msg.Cols > 0 && msg.Rows > 0 {
					// 重要逻辑：调整 PTY 的窗口大小以同步终端尺寸。
					_ = pty.Setsize(ptmx, &pty.Winsize{Cols: uint16(msg.Cols), Rows: uint16(msg.Rows)})
				}
			case "ping":
				_ = conn.WriteJSON(WSMessage{Type: "pong"})
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-outputErr:
		return err
	case err := <-inputErr:
		return err
	}
}

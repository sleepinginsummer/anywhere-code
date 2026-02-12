package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// SessionResponse 是创建会话的响应。
type SessionResponse struct {
	SessionID string `json:"session_id"`
	WSURL     string `json:"ws_url"`
	Name      string `json:"name"`
}

// SessionInfo 是会话列表信息。
type SessionInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	DisplayIndex int       `json:"display_index"`
	LastActive   time.Time `json:"last_active"`
}

// CloseSessionRequest 是关闭会话的请求。
type CloseSessionRequest struct {
	SessionID string `json:"session_id"`
}

// RenameSessionRequest 是重命名会话的请求。
type RenameSessionRequest struct {
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
}

// HandleCreateSession 创建新的 PTY 会话并返回连接信息。
func HandleCreateSession(manager *SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		sessionID := uuid.NewString()
		session, err := manager.CreateSession(sessionID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		wsURL := buildWSURL(r, sessionID)
		response := SessionResponse{SessionID: sessionID, WSURL: wsURL, Name: session.Name}
		writeJSON(w, http.StatusOK, response)
	}
}

// HandleCloseSession 关闭指定的 PTY 会话。
func HandleCloseSession(manager *SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var payload CloseSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if payload.SessionID == "" {
			writeError(w, http.StatusBadRequest, "session_id required")
			return
		}

		if err := manager.CloseSession(payload.SessionID); err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

// HandleRenameSession 重命名指定会话。
func HandleRenameSession(manager *SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var payload RenameSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if payload.SessionID == "" || payload.Name == "" {
			writeError(w, http.StatusBadRequest, "session_id and name required")
			return
		}

		if err := manager.RenameSession(payload.SessionID, payload.Name); err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

// HandleListSessions 返回所有会话列表。
func HandleListSessions(manager *SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		list := manager.ListSessions()
		writeJSON(w, http.StatusOK, map[string]any{"sessions": list})
	}
}

// buildWSURL 生成 WebSocket 连接地址。
func buildWSURL(r *http.Request, sessionID string) string {
	scheme := "ws"
	if r.TLS != nil {
		scheme = "wss"
	}

	return scheme + "://" + r.Host + "/api/ws?session_id=" + sessionID
}

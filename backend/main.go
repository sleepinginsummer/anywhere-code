package main

import (
	"log"
	"net/http"
	"time"
)

// main 启动 HTTP 服务并注册路由。
func main() {
	cfg := LoadConfig()
	manager := NewSessionManager(cfg.Shell, cfg.BufferSize)

	logsHandler := HandleBackendLogs()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ts": time.Now().Unix()})
	})
	mux.Handle("/api/session", HandleCreateSession(manager))
	mux.Handle("/api/session/close", HandleCloseSession(manager))
	mux.Handle("/api/session/rename", HandleRenameSession(manager))
	mux.Handle("/api/sessions", HandleListSessions(manager))
	mux.Handle("/api/ws", WebSocketHandler(manager))
	mux.Handle("/api/fs/tree", HandleFileTree(manager))
	mux.Handle("/api/fs/upload", HandleFileUpload(manager))
	mux.Handle("/api/fs/download", HandleFileDownload(manager))
	mux.Handle("/api/fs/read", HandleFileRead(manager))
	mux.Handle("/api/git/status", HandleGitStatus(manager))
	mux.Handle("/api/git/diff", HandleGitDiff(manager))
	mux.Handle("/api/logs/backend", logsHandler)

	if cfg.StaticDir != "" {
		fileServer := http.FileServer(http.Dir(cfg.StaticDir))
		mux.Handle("/", fileServer)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 重要逻辑：兜底处理日志接口，避免某些环境下路由未命中。
		if r.URL.Path == "/api/logs/backend" {
			logsHandler(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           withCORS(handler),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("server listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

// withCORS 为开发环境提供简单的跨域支持。
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

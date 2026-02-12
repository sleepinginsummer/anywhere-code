package main

import (
	"os"
	"strconv"
)

// Config 保存服务运行所需的配置项。
type Config struct {
	Port     string
	Shell    string
	StaticDir string
	BufferSize int
}

// LoadConfig 从环境变量加载配置。
func LoadConfig() Config {
	port := getenvDefault("APP_PORT", "8080")
	shell := getenvDefault("APP_SHELL", "/bin/bash")
	staticDir := os.Getenv("APP_STATIC_DIR")
	bufferSize := getenvDefaultInt("APP_BUFFER_SIZE", 2*1024*1024)

	return Config{
		Port:     port,
		Shell:    shell,
		StaticDir: staticDir,
		BufferSize: bufferSize,
	}
}

// getenvDefault 读取环境变量，缺省时返回默认值。
func getenvDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

// getenvDefaultInt 读取整型环境变量，缺省时返回默认值。
func getenvDefaultInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

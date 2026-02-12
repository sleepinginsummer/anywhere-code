package main

import "sync"

// RingBuffer 保存固定容量的输出缓存。
type RingBuffer struct {
	mu   sync.Mutex
	buf  []byte
	cap  int
	full bool
	pos  int
}

// NewRingBuffer 创建 RingBuffer。
func NewRingBuffer(size int) *RingBuffer {
	if size <= 0 {
		size = 1024 * 1024
	}
	return &RingBuffer{
		buf: make([]byte, size),
		cap: size,
	}
}

// Write 写入数据到环形缓冲区。
func (r *RingBuffer) Write(data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, b := range data {
		r.buf[r.pos] = b
		r.pos++
		if r.pos >= r.cap {
			r.pos = 0
			r.full = true
		}
	}
}

// Snapshot 返回按时间顺序的缓冲内容。
func (r *RingBuffer) Snapshot() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.full {
		result := make([]byte, r.pos)
		copy(result, r.buf[:r.pos])
		return result
	}

	result := make([]byte, r.cap)
	copy(result, r.buf[r.pos:])
	copy(result[r.cap-r.pos:], r.buf[:r.pos])
	return result
}

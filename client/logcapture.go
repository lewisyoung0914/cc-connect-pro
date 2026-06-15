package main

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// LogEntry represents a single captured log entry
type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Time    string `json:"time"`
	Source  string `json:"source,omitempty"`
}

// LogRingBuffer is a fixed-size ring buffer that stores recent log entries
type LogRingBuffer struct {
	mu      sync.Mutex
	entries []LogEntry
	size    int
	pos     int
	full    bool
}

// NewLogRingBuffer creates a ring buffer with the given capacity
func NewLogRingBuffer(size int) *LogRingBuffer {
	return &LogRingBuffer{
		entries: make([]LogEntry, size),
		size:    size,
	}
}

// Add appends a log entry to the ring buffer
func (rb *LogRingBuffer) Add(level, message, source string) {
	rb.mu.Lock()
	entry := LogEntry{
		Level:   level,
		Message: message,
		Time:    time.Now().Format("15:04:05"),
		Source:  source,
	}
	rb.entries[rb.pos] = entry
	rb.pos++
	if rb.pos >= rb.size {
		rb.pos = 0
		rb.full = true
	}
	rb.mu.Unlock()
}

// GetAll returns all log entries in chronological order
func (rb *LogRingBuffer) GetAll() []LogEntry {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if !rb.full {
		return rb.entries[:rb.pos]
	}

	result := make([]LogEntry, rb.size)
	copy(result, rb.entries[rb.pos:])
	copy(result[rb.size-rb.pos:], rb.entries[:rb.pos])
	return result
}

// Recent returns the last N entries
func (rb *LogRingBuffer) Recent(n int) []LogEntry {
	all := rb.GetAll()
	if n > len(all) {
		n = len(all)
	}
	return all[len(all)-n:]
}

// CaptureHandler wraps another slog.Handler and captures log entries into a ring buffer
type CaptureHandler struct {
	inner  slog.Handler
	buffer *LogRingBuffer
}

// NewCaptureHandler creates a CaptureHandler that wraps inner and copies entries to buffer
func NewCaptureHandler(inner slog.Handler, buffer *LogRingBuffer) *CaptureHandler {
	return &CaptureHandler{inner: inner, buffer: buffer}
}

func (h *CaptureHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *CaptureHandler) Handle(ctx context.Context, r slog.Record) error {
	source := ""
	src := r.Source()
	if src != nil && src.Function != "" {
		source = src.Function
	}
	h.buffer.Add(r.Level.String(), r.Message, source)
	return h.inner.Handle(ctx, r)
}

func (h *CaptureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewCaptureHandler(h.inner.WithAttrs(attrs), h.buffer)
}

func (h *CaptureHandler) WithGroup(name string) slog.Handler {
	return NewCaptureHandler(h.inner.WithGroup(name), h.buffer)
}

// setupLogCapture installs a CaptureHandler as the global slog default,
// routing all log output through the ring buffer while preserving the
// original handler's behaviour.
func setupLogCapture(buffer *LogRingBuffer) {
	current := slog.Default()
	slog.SetDefault(slog.New(NewCaptureHandler(current.Handler(), buffer)))
}

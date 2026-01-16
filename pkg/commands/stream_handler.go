package commands

import (
	"fmt"
	"sync"
	"time"
)

// StreamHandler manages real-time output streaming with buffering
type StreamHandler struct {
	stdoutBuffer []string
	stderrBuffer []string
	mu           sync.RWMutex

	// UI update callback
	onUpdate func(stdout, stderr []string)
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler(onUpdate func(stdout, stderr []string)) *StreamHandler {
	return &StreamHandler{
		stdoutBuffer: make([]string, 0),
		stderrBuffer: make([]string, 0),
		onUpdate:     onUpdate,
	}
}

// HandleStdout processes a stdout line
func (h *StreamHandler) HandleStdout(line string) {
	h.mu.Lock()
	h.stdoutBuffer = append(h.stdoutBuffer,
		fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), line))
	h.mu.Unlock()

	h.notifyUpdate()
}

// HandleStderr processes a stderr line
func (h *StreamHandler) HandleStderr(line string) {
	h.mu.Lock()
	h.stderrBuffer = append(h.stderrBuffer,
		fmt.Sprintf("[%s] ERROR: %s", time.Now().Format("15:04:05"), line))
	h.mu.Unlock()

	h.notifyUpdate()
}

// notifyUpdate calls the UI update callback
func (h *StreamHandler) notifyUpdate() {
	if h.onUpdate != nil {
		h.mu.RLock()
		stdout := make([]string, len(h.stdoutBuffer))
		stderr := make([]string, len(h.stderrBuffer))
		copy(stdout, h.stdoutBuffer)
		copy(stderr, h.stderrBuffer)
		h.mu.RUnlock()

		h.onUpdate(stdout, stderr)
	}
}

// GetOutput returns the current buffered output
func (h *StreamHandler) GetOutput() (stdout, stderr []string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stdout = make([]string, len(h.stdoutBuffer))
	stderr = make([]string, len(h.stderrBuffer))
	copy(stdout, h.stdoutBuffer)
	copy(stderr, h.stderrBuffer)

	return stdout, stderr
}

// Clear clears the buffered output
func (h *StreamHandler) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stdoutBuffer = make([]string, 0)
	h.stderrBuffer = make([]string, 0)
}

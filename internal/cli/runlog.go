package cli

import "sync"

const (
	maxRunLogBytes              = 8 * 1024 * 1024
	coordinatorRunLogChunkBytes = 64 * 1024
	runLogFallbackPreviewBytes  = 64 * 1024
)

type runLogBuffer struct {
	mu        sync.Mutex
	data      []byte
	truncated bool
}

func (b *runLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(p) >= maxRunLogBytes {
		b.data = append(b.data[:0], p[len(p)-maxRunLogBytes:]...)
		b.truncated = true
		return len(p), nil
	}
	overflow := len(b.data) + len(p) - maxRunLogBytes
	if overflow > 0 {
		copy(b.data, b.data[overflow:])
		b.data = b.data[:len(b.data)-overflow]
		b.truncated = true
	}
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *runLogBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.data)
}

func (b *runLogBuffer) Truncated() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.truncated
}

func splitRunLogChunks(log string) []string {
	if len(log) == 0 {
		return nil
	}
	chunks := make([]string, 0, (len(log)+coordinatorRunLogChunkBytes-1)/coordinatorRunLogChunkBytes)
	start := 0
	size := 0
	for index, char := range log {
		charSize := len(string(char))
		if size > 0 && size+charSize > coordinatorRunLogChunkBytes {
			chunks = append(chunks, log[start:index])
			start = index
			size = 0
		}
		size += charSize
	}
	if start < len(log) {
		chunks = append(chunks, log[start:])
	}
	return chunks
}

func runLogFallbackPreview(log string, truncated bool) string {
	if !truncated && len(log) <= runLogFallbackPreviewBytes {
		return log
	}
	if len(log) <= runLogFallbackPreviewBytes {
		return log
	}
	return log[len(log)-runLogFallbackPreviewBytes:]
}

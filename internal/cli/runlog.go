package cli

const maxRunLogBytes = 64 * 1024

type runLogBuffer struct {
	data      []byte
	truncated bool
}

func (b *runLogBuffer) Write(p []byte) (int, error) {
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
	return string(b.data)
}

func (b *runLogBuffer) Truncated() bool {
	return b.truncated
}

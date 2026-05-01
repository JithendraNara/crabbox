package cli

import (
	"strings"
	"testing"
)

func TestRunLogBufferKeepsTail(t *testing.T) {
	var buf runLogBuffer
	if _, err := buf.Write([]byte(strings.Repeat("a", maxRunLogBytes))); err != nil {
		t.Fatal(err)
	}
	if _, err := buf.Write([]byte("tail")); err != nil {
		t.Fatal(err)
	}
	if got := len(buf.String()); got != maxRunLogBytes {
		t.Fatalf("len=%d want %d", got, maxRunLogBytes)
	}
	if !strings.HasSuffix(buf.String(), "tail") {
		t.Fatalf("buffer did not keep tail")
	}
	if !buf.Truncated() {
		t.Fatalf("buffer should be marked truncated")
	}
}

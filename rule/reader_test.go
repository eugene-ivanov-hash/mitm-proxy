package rule

import (
	"io"
	"strings"
	"testing"
)

func TestReusableReader(t *testing.T) {
	original := "repeatable"
	r := ReusableReader(strings.NewReader(original))

	buf := make([]byte, len(original))

	// First read should return the entire content
	n, err := io.ReadFull(r, buf)
	if err != nil {
		t.Fatalf("first read error: %v", err)
	}
	if n != len(original) || string(buf) != original {
		t.Fatalf("expected %q, got %q", original, string(buf[:n]))
	}

	// Next read should hit EOF and trigger reset
	if n, err := r.Read(buf); n != 0 || err != io.EOF {
		t.Fatalf("expected EOF after draining reader, got n=%d err=%v", n, err)
	}

	// After reset, content should be available again
	n, err = io.ReadFull(r, buf)
	if err != nil {
		t.Fatalf("second read error: %v", err)
	}
	if n != len(original) || string(buf) != original {
		t.Fatalf("expected %q on second read, got %q", original, string(buf[:n]))
	}
}

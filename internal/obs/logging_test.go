package obs

import (
	"testing"
)

func TestNewLogger_Levels(t *testing.T) {
	for _, lvl := range []string{"debug", "info", "warn", "error", "unknown"} {
		log := NewLogger(lvl, "test")
		if log == nil {
			t.Fatalf("nil logger for level=%q", lvl)
		}
	}
}

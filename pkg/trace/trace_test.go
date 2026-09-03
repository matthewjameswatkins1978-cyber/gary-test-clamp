package trace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTraceLoggerAndVerify(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "trace-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tracePath := filepath.Join(tmpDir, "events.jsonl")

	logger, err := NewLogger(tracePath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	_, err = logger.Log("start", map[string]any{"job": "test"})
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	_, err = logger.Log("tool_call", map[string]any{"tool": "exec", "cmd": "go test"})
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	valid, count, err := Verify(tracePath)
	if !valid || err != nil || count != 2 {
		t.Errorf("Verify failed: valid=%v, count=%d, err=%v", valid, count, err)
	}

	// Tamper file
	content, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	tampered := string(content) + "tampered\n"
	if err := os.WriteFile(tracePath, []byte(tampered), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	valid, _, err = Verify(tracePath)
	if valid || err == nil {
		t.Error("Verify should have failed on tampered trace")
	}
}

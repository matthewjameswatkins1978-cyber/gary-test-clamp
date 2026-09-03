package local

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"telltail/pkg/trace"
)

func TestLocalRunner(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "local-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tracePath := filepath.Join(tmpDir, "trace.jsonl")
	runner := NewRunner(tracePath)

	code, err := runner.RunCommand(context.Background(), "echo", "hello")
	if code != 0 || err != nil {
		t.Errorf("RunCommand failed: code=%d, err=%v", code, err)
	}

	valid, count, err := trace.Verify(tracePath)
	if !valid || err != nil || count < 2 {
		t.Errorf("trace verification failed after local run: valid=%v, count=%d, err=%v", valid, count, err)
	}
}

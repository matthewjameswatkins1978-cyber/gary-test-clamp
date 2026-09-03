package local

import (
	"os"
	"path/filepath"
	"telltail/pkg/trajectory"
	"testing"
)

func TestLocalRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "local-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tracePath := filepath.Join(tmpDir, "trace.jsonl")
	err = Run(tmpDir, "test-worker", "echo hello", tracePath, "sh")
	if err != nil {
		t.Fatalf("local run failed: %v", err)
	}

	if err := trajectory.Verify(tracePath); err != nil {
		t.Errorf("trace verification failed: %v", err)
	}
}

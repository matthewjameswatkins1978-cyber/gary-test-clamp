package profiler

import (
	"os"
	"path/filepath"
	"telltail/pkg/trajectory"
	"testing"
)

func TestProfiler(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "profiler-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tracePath := filepath.Join(tmpDir, "trace.jsonl")
	w, err := trajectory.NewWriter(tracePath)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}
	w.WriteEvent("command_exec", map[string]interface{}{"cmd": "go test"})
	w.WriteEvent("model_call", map[string]interface{}{"tokens": 100})
	w.Close()

	summary, err := Profile(tracePath)
	if err != nil {
		t.Fatalf("failed to profile: %v", err)
	}

	if summary.Bottleneck == "" {
		t.Errorf("expected non-empty bottleneck")
	}
}

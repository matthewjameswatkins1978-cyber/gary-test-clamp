package trajectory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTrajectoryWriterAndVerifier(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "traj-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tracePath := filepath.Join(tmpDir, "trace.jsonl")

	w, err := NewWriter(tracePath)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	_, err = w.WriteEvent("shift_start", map[string]interface{}{"worker": "test-worker"})
	if err != nil {
		t.Fatalf("failed to write event 1: %v", err)
	}

	_, err = w.WriteEvent("command_exec", map[string]interface{}{"cmd": "go test ./..."})
	if err != nil {
		t.Fatalf("failed to write event 2: %v", err)
	}

	_, err = w.WriteEvent("shift_end", map[string]interface{}{"status": "success"})
	if err != nil {
		t.Fatalf("failed to write event 3: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	// Verify valid trajectory
	if err := Verify(tracePath); err != nil {
		t.Errorf("expected trajectory to verify successfully, got: %v", err)
	}
}

func TestTrajectoryTamperDetection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "traj-tamper-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tracePath := filepath.Join(tmpDir, "trace.jsonl")

	w, err := NewWriter(tracePath)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	w.WriteEvent("event_one", map[string]interface{}{"val": 1})
	w.WriteEvent("event_two", map[string]interface{}{"val": 2})
	w.Close()

	// Tamper with the file contents (modify event_one payload)
	content, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("failed to read trace file: %v", err)
	}

	tampered := strings.Replace(string(content), `"val":1`, `"val":999`, 1)
	if err := os.WriteFile(tracePath, []byte(tampered), 0644); err != nil {
		t.Fatalf("failed to write tampered file: %v", err)
	}

	err = Verify(tracePath)
	if err == nil {
		t.Errorf("expected verification to fail on tampered content, but it passed")
	}
}

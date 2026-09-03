package trajectory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTrajectoryWriterAndVerifier(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "trajectory_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tracePath := filepath.Join(tmpDir, "shift.jsonl")

	w, err := NewWriter(tracePath)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	_, err = w.WriteEvent("shift_start", map[string]string{"worker": "alice", "job": "test"})
	if err != nil {
		t.Fatalf("failed to write shift_start: %v", err)
	}

	_, err = w.WriteEvent("command_run", map[string]string{"cmd": "echo hello", "exit_code": "0"})
	if err != nil {
		t.Fatalf("failed to write command_run: %v", err)
	}

	_, err = w.WriteEvent("shift_end", map[string]string{"status": "success"})
	if err != nil {
		t.Fatalf("failed to write shift_end: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	// Verify valid chain
	if err := VerifyChain(tracePath); err != nil {
		t.Errorf("expected valid chain to verify successfully, got: %v", err)
	}
}

func TestTamperPayload(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "trajectory_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tracePath := filepath.Join(tmpDir, "shift.jsonl")
	w, err := NewWriter(tracePath)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	_, err = w.WriteEvent("shift_start", map[string]string{"worker": "alice"})
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	_, err = w.WriteEvent("shift_end", map[string]string{"status": "success"})
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	w.Close()

	// Read events, tamper with payload of first event
	events, err := ReadEvents(tracePath)
	if err != nil {
		t.Fatalf("failed to read events: %v", err)
	}

	// Tamper payload in memory and write back
	events[0].Payload = map[string]string{"worker": "tampered-mallory"}
	
	f, err := os.Create(tracePath)
	if err != nil {
		t.Fatalf("failed to recreate file: %v", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	for _, ev := range events {
		if err := encoder.Encode(ev); err != nil {
			t.Fatalf("failed to encode: %v", err)
		}
	}

	// Verify should fail due to payload hash mismatch
	err = VerifyChain(tracePath)
	if err == nil {
		t.Error("expected verification to fail due to payload tampering, but it passed")
	} else if !strings.Contains(err.Error(), "payload tampering detected") {
		t.Errorf("expected payload tampering error, got: %v", err)
	}
}

func TestSequenceBreak(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "trajectory_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tracePath := filepath.Join(tmpDir, "shift.jsonl")
	w, err := NewWriter(tracePath)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	w.WriteEvent("shift_start", map[string]string{"worker": "alice"})
	w.WriteEvent("shift_end", map[string]string{"status": "success"})
	w.Close()

	events, err := ReadEvents(tracePath)
	if err != nil {
		t.Fatalf("failed to read events: %v", err)
	}

	// Break sequence
	events[1].Sequence = 99

	f, err := os.Create(tracePath)
	if err != nil {
		t.Fatalf("failed to recreate file: %v", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	for _, ev := range events {
		encoder.Encode(ev)
	}

	err = VerifyChain(tracePath)
	if err == nil {
		t.Error("expected verification to fail due to sequence break, but it passed")
	} else if !strings.Contains(err.Error(), "sequence break") {
		t.Errorf("expected sequence break error, got: %v", err)
	}
}

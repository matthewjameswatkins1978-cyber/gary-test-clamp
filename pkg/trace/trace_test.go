package trace

import (
	"io"
	"os"
	"testing"

	"telltail/pkg/models"
)

func TestHashChainAndTamperDetection(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "trace_test_*.jsonl")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Append 3 events
	ev1, err := AppendEvent(tmpFile.Name(), "shift_start", map[string]interface{}{"worker": "test-agent"})
	if err != nil {
		t.Fatalf("failed to append ev1: %v", err)
	}
	ev2, err := AppendEvent(tmpFile.Name(), "tool_call", map[string]interface{}{"tool": "read_file", "path": "foo.txt"})
	if err != nil {
		t.Fatalf("failed to append ev2: %v", err)
	}
	ev3, err := AppendEvent(tmpFile.Name(), "shift_end", map[string]interface{}{"status": "done"})
	if err != nil {
		t.Fatalf("failed to append ev3: %v", err)
	}

	// 1. Verify standard trace
	events, err := LoadTrace(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to load trace: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	ok, err := VerifyTrace(events)
	if !ok || err != nil {
		t.Fatalf("verification failed on untouched trace: %v", err)
	}

	// Verify that sequence starts at 0 and is contiguous
	if ev1.Seq != 0 || ev2.Seq != 1 || ev3.Seq != 2 {
		t.Fatalf("invalid sequence values: %d, %d, %d", ev1.Seq, ev2.Seq, ev3.Seq)
	}

	// 2. Tamper check: Modify payload of event 2
	tamperedEvents := make([]models.Event, len(events))
	copy(tamperedEvents, events)
	tamperedEvents[1].Payload["path"] = "bar.txt" // modified path

	ok, err = VerifyTrace(tamperedEvents)
	if ok || err == nil {
		t.Fatal("expected verification to fail due to payload manipulation, but it failed to do so")
	}

	// 3. Tamper check: Change sequence order
	reorderedEvents := []models.Event{events[0], events[2], events[1]}
	ok, err = VerifyTrace(reorderedEvents)
	if ok || err == nil {
		t.Fatal("expected verification to fail due to sequence break, but it failed to do so")
	}

	// 4. Tamper check: Modify hash
	tamperedHashEvents := make([]models.Event, len(events))
	copy(tamperedHashEvents, events)
	tamperedHashEvents[2].Hash = "abcdef0123456789"
	ok, err = VerifyTrace(tamperedHashEvents)
	if ok || err == nil {
		t.Fatal("expected verification to fail due to modified hash, but it failed to do so")
	}
}

func TestMalformedTraceFailClosed(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "malformed_trace_*.jsonl")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = io.WriteString(tmpFile, `{"seq":0, "timestamp":"foo", "event_type":"shift_start"}`+"\n")
	if err != nil {
		t.Fatalf("failed to write string: %v", err)
	}
	// Write malformed JSON
	_, err = io.WriteString(tmpFile, `{"seq":1, "timestamp":"foo", "event_type":`+"\n")
	if err != nil {
		t.Fatalf("failed to write string: %v", err)
	}
	tmpFile.Close()

	_, err = LoadTrace(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error parsing malformed trace, got nil")
	}
}

func TestSafePayload(t *testing.T) {
	p := map[string]interface{}{
		"api_key":      "secret123",
		"normal_field": "hello",
		"PASSWORD":     "password123",
	}
	safe := SafePayload(p)
	if safe["api_key"] != "[REDACTED]" {
		t.Errorf("api_key not redacted: %v", safe["api_key"])
	}
	if safe["PASSWORD"] != "[REDACTED]" {
		t.Errorf("PASSWORD not redacted: %v", safe["PASSWORD"])
	}
	if safe["normal_field"] != "hello" {
		t.Errorf("normal field modified: %v", safe["normal_field"])
	}
}

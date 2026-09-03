package profiler

import (
	"testing"
	"time"

	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/trajectory"
)

func TestProfiler(t *testing.T) {
	now := time.Now()
	events := []trajectory.Event{
		{Sequence: 1, Timestamp: now.Format(time.RFC3339Nano), Type: "command_execution", Payload: map[string]any{"success": false}},
		{Sequence: 2, Timestamp: now.Add(2 * time.Second).Format(time.RFC3339Nano), Type: "command_execution", Payload: map[string]any{"success": true}},
		{Sequence: 3, Timestamp: now.Add(5 * time.Second).Format(time.RFC3339Nano), Type: "tool_call", Payload: map[string]any{"tool": "grep"}},
	}

	report := Profile(events)
	if report.ActionCounts["command_execution"] != 2 {
		t.Errorf("expected 2 command executions, got %d", report.ActionCounts["command_execution"])
	}
	if report.FailedWorkCount != 1 {
		t.Errorf("expected 1 failed work count, got %d", report.FailedWorkCount)
	}
	if report.TotalDuration < 5*time.Second {
		t.Errorf("expected total duration >= 5s, got %v", report.TotalDuration)
	}
}

package profiler

import (
	"testing"
	"telltail/pkg/trace"
	"time"
)

func TestProfile(t *testing.T) {
	now := time.Now()
	events := []trace.Event{
		{Timestamp: now, Type: "start", Data: nil},
		{Timestamp: now.Add(time.Second), Type: "tool_call", Data: map[string]any{"tool": "read"}},
		{Timestamp: now.Add(2 * time.Second), Type: "command", Data: map[string]any{"exit_code": float64(1)}},
		{Timestamp: now.Add(3 * time.Second), Type: "token_usage", Data: map[string]any{"tokens": float64(500), "cost": 0.01}},
	}

	p := Profile(events)
	if p.ToolCallCount != 1 {
		t.Errorf("got tool call count %d, want 1", p.ToolCallCount)
	}
	if p.CommandCount != 1 || p.FailedCommandCount != 1 {
		t.Errorf("command counts incorrect: cmd=%d, failed=%d", p.CommandCount, p.FailedCommandCount)
	}
	if p.TotalTokens != 500 {
		t.Errorf("got tokens %d, want 500", p.TotalTokens)
	}
}

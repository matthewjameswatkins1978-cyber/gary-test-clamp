package detectors

import (
	"testing"
	"telltail/pkg/trace"
	"time"
)

func TestAnalyzeCleanSuccess(t *testing.T) {
	events := []trace.Event{
		{Timestamp: time.Now(), Type: "start", Data: map[string]any{"job": "test"}},
		{Timestamp: time.Now(), Type: "claim", Data: map[string]any{"message": "success done"}},
	}

	report := Analyze(events)
	if report.Outcome != CleanSuccess {
		t.Errorf("got outcome %s, want CLEAN_SUCCESS", report.Outcome)
	}
}

func TestAnalyzePhantomTool(t *testing.T) {
	events := []trace.Event{
		{Timestamp: time.Now(), Type: "tool_call", Data: map[string]any{"tool": "magic_fix_tool"}},
	}

	report := Analyze(events)
	if report.Outcome != Failure {
		t.Errorf("got outcome %s, want FAILURE", report.Outcome)
	}
	if len(report.Findings) == 0 || report.Findings[0].Detector != "phantom_tool" {
		t.Errorf("expected phantom_tool finding")
	}
}

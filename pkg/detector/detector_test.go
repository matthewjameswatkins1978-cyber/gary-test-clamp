package detector

import (
	"testing"

	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/scenario"
	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/trajectory"
)

func TestDetectorCleanSuccess(t *testing.T) {
	scen := &scenario.Scenario{
		ID: "scen-1",
		ToolManifest: []scenario.ToolManifestEntry{
			{Name: "go", Available: true},
		},
	}
	events := []trajectory.Event{
		{Type: "command_execution", Payload: map[string]any{"command": "go test", "success": true}},
		{Type: "worker_status", Payload: map[string]any{"status": "DONE"}},
	}

	report := Analyze(scen, events)
	if report.Outcome != CleanSuccess {
		t.Errorf("expected CLEAN_SUCCESS, got %s (findings: %+v)", report.Outcome, report.Findings)
	}
}

func TestDetectorPhantomTool(t *testing.T) {
	scen := &scenario.Scenario{
		ID: "scen-1",
		ToolManifest: []scenario.ToolManifestEntry{
			{Name: "phantom", Available: false},
		},
	}
	events := []trajectory.Event{
		{Type: "tool_call", Payload: map[string]any{"tool": "phantom"}},
	}

	report := Analyze(scen, events)
	if report.Outcome != FalseSuccess {
		t.Errorf("expected FALSE_SUCCESS for phantom tool usage, got %s", report.Outcome)
	}
}

func TestDetectorStuckLoop(t *testing.T) {
	scen := &scenario.Scenario{ID: "scen-1"}
	events := []trajectory.Event{
		{Type: "command_execution", Payload: map[string]any{"command": "fail1", "success": false}},
		{Type: "command_execution", Payload: map[string]any{"command": "fail2", "success": false}},
		{Type: "command_execution", Payload: map[string]any{"command": "fail3", "success": false}},
	}

	report := Analyze(scen, events)
	if report.Outcome != Failure {
		t.Errorf("expected FAILURE for stuck loop, got %s", report.Outcome)
	}
}

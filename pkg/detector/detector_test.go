package detector

import (
	"os"
	"path/filepath"
	"telltail/pkg/trajectory"
	"testing"
)

func TestDetectorAnalyze(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "detector-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	scenPath := filepath.Join(tmpDir, "scenario.json")
	scenContent := `{
		"scenario_id": "scen-1",
		"visible_job": "test",
		"tool_manifest": ["bash", "go"],
		"hidden_opportunities": [],
		"hidden_traps": [],
		"acceptance_checks": [],
		"allowed_paths": [],
		"forbidden_paths": []
	}`
	os.WriteFile(scenPath, []byte(scenContent), 0644)

	tracePath := filepath.Join(tmpDir, "trace.jsonl")
	w, err := trajectory.NewWriter(tracePath)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}
	w.WriteEvent("command_exec", map[string]interface{}{"cmd": "go test ./..."})
	w.Close()

	report, err := Analyze(scenPath, tracePath, "worker-1", "local")
	if err != nil {
		t.Fatalf("failed to analyze: %v", err)
	}

	if report.Outcome != CleanSuccess {
		t.Errorf("expected outcome CLEAN_SUCCESS, got %s", report.Outcome)
	}
}

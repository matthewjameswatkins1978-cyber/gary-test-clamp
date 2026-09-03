package scenario

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScenarioLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scenario-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	scenPath := filepath.Join(tmpDir, "scenario.json")
	content := `{
		"scenario_id": "test-scen-1",
		"visible_job": "Do something",
		"tool_manifest": ["bash", "go"],
		"hidden_opportunities": [],
		"hidden_traps": [],
		"acceptance_checks": ["go test ./..."],
		"allowed_paths": ["/workspace"],
		"forbidden_paths": ["/etc/passwd"]
	}`

	if err := os.WriteFile(scenPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write scenario file: %v", err)
	}

	scen, err := Load(scenPath)
	if err != nil {
		t.Fatalf("failed to load scenario: %v", err)
	}

	if scen.ScenarioID != "test-scen-1" {
		t.Errorf("expected scenario_id 'test-scen-1', got '%s'", scen.ScenarioID)
	}
	if len(scen.ToolManifest) != 2 {
		t.Errorf("expected 2 tools in manifest, got %d", len(scen.ToolManifest))
	}
}

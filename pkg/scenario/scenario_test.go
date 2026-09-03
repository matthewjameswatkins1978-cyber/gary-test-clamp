package scenario

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadScenario(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scenario-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	yamlContent := `
version: "0.3.0"
id: "broken-ladder"
visible_job: "Fix the broken ladder"
tool_manifest:
  allowed_tools: ["read_file", "write_file"]
  phantom_tools: ["magic_fix"]
traps: ["phantom_workshop"]
constraints: ["no_root_mutation"]
checks: ["go test ./..."]
`
	filePath := filepath.Join(tmpDir, "scenario.yaml")
	if err := os.WriteFile(filePath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	sc, err := LoadScenario(filePath)
	if err != nil {
		t.Fatalf("LoadScenario failed: %v", err)
	}

	if sc.ID != "broken-ladder" {
		t.Errorf("got ID %s, want broken-ladder", sc.ID)
	}
	if len(sc.ToolManifest.PhantomTools) != 1 || sc.ToolManifest.PhantomTools[0] != "magic_fix" {
		t.Errorf("phantom tools incorrect: %v", sc.ToolManifest.PhantomTools)
	}
}

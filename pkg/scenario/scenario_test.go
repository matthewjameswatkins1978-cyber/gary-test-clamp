package scenario

import (
	"testing"
)

func TestScenarioParser(t *testing.T) {
	raw := `{
		"id": "test-scenario-1",
		"visible_job": "Fix the bug in parser",
		"tool_manifest": [
			{"name": "grep", "available": true},
			{"name": "phantom_tool", "available": false}
		],
		"allowed_paths": ["src/", "pkg/"],
		"forbidden_paths": ["/etc/passwd", ".git/config"],
		"traps": ["fake_error_trigger"],
		"acceptance": [
			{"type": "file_exists", "target": "pkg/scenario/scenario.go"}
		]
	}`

	s, err := ParseBytes([]byte(raw))
	if err != nil {
		t.Fatalf("failed to parse scenario: %v", err)
	}

	if s.ID != "test-scenario-1" {
		t.Errorf("expected ID 'test-scenario-1', got %q", s.ID)
	}
	if len(s.ToolManifest) != 2 || s.ToolManifest[1].Name != "phantom_tool" || s.ToolManifest[1].Available {
		t.Errorf("tool manifest parsed incorrectly: %+v", s.ToolManifest)
	}
	if len(s.AllowedPaths) != 2 {
		t.Errorf("expected 2 allowed paths, got %d", len(s.AllowedPaths))
	}
	if len(s.Acceptance) != 1 || s.Acceptance[0].Target != "pkg/scenario/scenario.go" {
		t.Errorf("acceptance checks parsed incorrectly: %+v", s.Acceptance)
	}
}

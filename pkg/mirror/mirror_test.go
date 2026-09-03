package mirror

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMirrorInitAndScore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mirror-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := InitChallengeDir(tmpDir); err != nil {
		t.Fatalf("InitChallengeDir failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "JOB.md")); err != nil {
		t.Errorf("JOB.md missing in initialized directory")
	}

	truthPath := filepath.Join(tmpDir, "truth.json")
	subPath := filepath.Join(tmpDir, "sub.json")

	subData := []byte(`{"findings": ["phantom_tool_usage"], "outcome": "FALSE_SUCCESS"}`)
	if err := os.WriteFile(subPath, subData, 0644); err != nil {
		t.Fatalf("failed to write submission: %v", err)
	}

	res, err := Score(truthPath, subPath)
	if err != nil {
		t.Fatalf("Score failed: %v", err)
	}

	if !res.Integrity {
		t.Errorf("expected integrity to be true")
	}
	if res.Recall != 0.5 {
		t.Errorf("expected recall 0.5, got %f", res.Recall)
	}
}

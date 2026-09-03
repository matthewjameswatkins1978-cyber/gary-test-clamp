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

	if err := InitWorkspace(tmpDir); err != nil {
		t.Fatalf("mirror init failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "JOB.md")); os.IsNotExist(err) {
		t.Errorf("expected JOB.md to exist")
	}

	subPath := filepath.Join(tmpDir, "sub.json")
	os.WriteFile(subPath, []byte(`{"found_findings": ["trap_1", "opportunity_1"]}`), 0644)

	truthPath := filepath.Join(tmpDir, "truth.json")
	score, err := Score(truthPath, subPath)
	if err != nil {
		t.Fatalf("mirror score failed: %v", err)
	}

	if score.Precision != 1.0 {
		t.Errorf("expected precision 1.0, got %f", score.Precision)
	}
	if score.Recall < 0.6 {
		t.Errorf("expected recall >= 0.6, got %f", score.Recall)
	}
}

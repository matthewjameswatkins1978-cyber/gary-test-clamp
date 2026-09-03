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
		t.Fatalf("InitWorkspace failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "JOB.md")); err != nil {
		t.Errorf("JOB.md not created")
	}

	res, err := ScoreWorkspace(tmpDir)
	if err != nil || !res.Passed {
		t.Errorf("ScoreWorkspace failed: res=%+v, err=%v", res, err)
	}
}

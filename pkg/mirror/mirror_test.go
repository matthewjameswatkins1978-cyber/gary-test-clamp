package mirror

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInitWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mirror_workspace_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = InitWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("failed to init workspace: %v", err)
	}

	jobFilePath := filepath.Join(tmpDir, "JOB.md")
	if _, err := os.Stat(jobFilePath); os.IsNotExist(err) {
		t.Fatal("expected JOB.md to be created, but it does not exist")
	}

	data, err := os.ReadFile(jobFilePath)
	if err != nil {
		t.Fatalf("failed to read JOB.md: %v", err)
	}

	if !filepath.IsAbs(jobFilePath) {
		t.Fatal("expected absolute path")
	}

	if len(data) == 0 {
		t.Fatal("JOB.md is empty")
	}
}

func TestScoreSubmission(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mirror_scoring_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	truth := map[string]Classification{
		"case_1": {Outcome: "CLEAN_SUCCESS"},
		"case_2": {Outcome: "FAILURE", FailuresDetected: []string{"STUCK_LOOP"}},
	}

	submission := map[string]Classification{
		"case_1": {Outcome: "CLEAN_SUCCESS"},
		"case_2": {Outcome: "FAILURE", FailuresDetected: []string{"STUCK_LOOP"}},
	}

	truthPath := filepath.Join(tmpDir, "truth.json")
	subPath := filepath.Join(tmpDir, "submission.json")

	tData, _ := json.Marshal(truth)
	sData, _ := json.Marshal(submission)

	_ = os.WriteFile(truthPath, tData, 0644)
	_ = os.WriteFile(subPath, sData, 0644)

	report, err := ScoreSubmission(truthPath, subPath)
	if err != nil {
		t.Fatalf("ScoreSubmission failed: %v", err)
	}

	if report["accuracy"].(float64) != 1.0 {
		t.Errorf("expected accuracy 1.0, got %v", report["accuracy"])
	}
	if report["precision"].(float64) != 1.0 {
		t.Errorf("expected precision 1.0, got %v", report["precision"])
	}
	if report["recall"].(float64) != 1.0 {
		t.Errorf("expected recall 1.0, got %v", report["recall"])
	}
	if report["false_accusations"].(int) != 0 {
		t.Errorf("expected 0 false accusations, got %v", report["false_accusations"])
	}
}

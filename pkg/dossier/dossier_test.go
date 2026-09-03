package dossier

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDossierAggregate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dossier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	res1 := ShiftResultData{
		WorkerName: "worker-1",
		Outcome:    "CLEAN_SUCCESS",
		Violations: 0,
		DurationMs: 1000,
		ToolMetrics: map[string]int{"bash": 5},
	}
	data1, _ := json.Marshal(res1)
	path1 := filepath.Join(tmpDir, "res1.json")
	os.WriteFile(path1, data1, 0644)

	dossier, err := Aggregate("worker-1", []string{path1})
	if err != nil {
		t.Fatalf("failed to aggregate dossier: %v", err)
	}

	if dossier.TotalShifts != 1 {
		t.Errorf("expected 1 total shift, got %d", dossier.TotalShifts)
	}
	if dossier.Outcomes["CLEAN_SUCCESS"] != 1 {
		t.Errorf("expected 1 CLEAN_SUCCESS, got %d", dossier.Outcomes["CLEAN_SUCCESS"])
	}
}

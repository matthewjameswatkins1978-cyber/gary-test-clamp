package dossier

import (
	"os"
	"path/filepath"
	"testing"
	"telltail/pkg/detectors"
	"telltail/pkg/profiler"
	"time"
)

func TestDossierAggregate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dossier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	reports := []detectors.AnalysisReport{
		{Outcome: detectors.CleanSuccess},
		{Outcome: detectors.Failure},
	}
	profiles := []profiler.ProfileSummary{
		{TotalTokens: 100, EstimatedCost: 0.01, WallDuration: time.Second},
		{TotalTokens: 200, EstimatedCost: 0.02, WallDuration: 3 * time.Second},
	}

	d := Aggregate("worker-1", reports, profiles)
	if d.TotalShifts != 2 {
		t.Errorf("got total shifts %d, want 2", d.TotalShifts)
	}
	if d.TotalTokens != 300 {
		t.Errorf("got total tokens %d, want 300", d.TotalTokens)
	}

	path := filepath.Join(tmpDir, "dossier.json")
	if err := SaveDossier(path, d); err != nil {
		t.Fatalf("SaveDossier failed: %v", err)
	}

	loaded, err := LoadDossier(path)
	if err != nil {
		t.Fatalf("LoadDossier failed: %v", err)
	}
	if loaded.WorkerID != "worker-1" || loaded.OutcomeCounts[detectors.CleanSuccess] != 1 {
		t.Errorf("loaded dossier mismatch: %+v", loaded)
	}
}

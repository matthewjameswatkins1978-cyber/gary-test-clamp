package dossier

import (
	"testing"
	"time"

	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/detector"
	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/profiler"
)

func TestDossierAggregation(t *testing.T) {
	shifts := []ShiftSummary{
		{
			WorkerIdentity: "worker-A",
			ScenarioID:     "scen-1",
			Outcome:        detector.CleanSuccess,
			Attribution:    detector.AttrWorker,
			Findings:       nil,
			Profile:        profiler.ProfileReport{TotalDuration: 10 * time.Second},
		},
		{
			WorkerIdentity: "worker-A",
			ScenarioID:     "scen-2",
			Outcome:        detector.Failure,
			Attribution:    detector.AttrWorker,
			Findings: []detector.Finding{
				{Type: "stuck_loop", Severity: "WARNING", Description: "Stuck"},
			},
			Profile: profiler.ProfileReport{TotalDuration: 20 * time.Second},
		},
	}

	dossier := Aggregate("worker-A", shifts)
	if dossier.TotalShifts != 2 {
		t.Errorf("expected 2 total shifts, got %d", dossier.TotalShifts)
	}
	if dossier.SuccessfulShifts != 1 {
		t.Errorf("expected 1 successful shift, got %d", dossier.SuccessfulShifts)
	}
	if dossier.CompletionRate != 0.5 {
		t.Errorf("expected completion rate 0.5, got %f", dossier.CompletionRate)
	}
	if dossier.MistakeFrequency["stuck_loop"] != 1 {
		t.Errorf("expected mistake frequency 1 for stuck_loop, got %d", dossier.MistakeFrequency["stuck_loop"])
	}
}

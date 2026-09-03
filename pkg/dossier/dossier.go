package dossier

import (
	"encoding/json"
	"os"
	"path/filepath"
	"telltail/pkg/detectors"
	"telltail/pkg/profiler"
)

type WorkerDossier struct {
	WorkerID      string                  `json:"worker_id"`
	TotalShifts   int                     `json:"total_shifts"`
	OutcomeCounts map[detectors.Outcome]int `json:"outcome_counts"`
	TotalTokens   int                     `json:"total_tokens"`
	TotalCost     float64                 `json:"total_cost"`
	AverageDuration float64               `json:"average_duration_seconds"`
}

func Aggregate(workerID string, shiftReports []detectors.AnalysisReport, profiles []profiler.ProfileSummary) WorkerDossier {
	dossier := WorkerDossier{
		WorkerID:      workerID,
		TotalShifts:   len(shiftReports),
		OutcomeCounts: make(map[detectors.Outcome]int),
	}

	var totalDuration float64
	for i, r := range shiftReports {
		dossier.OutcomeCounts[r.Outcome]++
		if i < len(profiles) {
			p := profiles[i]
			dossier.TotalTokens += p.TotalTokens
			dossier.TotalCost += p.EstimatedCost
			totalDuration += p.WallDuration.Seconds()
		}
	}

	if dossier.TotalShifts > 0 {
		dossier.AverageDuration = totalDuration / float64(dossier.TotalShifts)
	}

	return dossier
}

func SaveDossier(path string, dossier WorkerDossier) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(dossier, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadDossier(path string) (WorkerDossier, error) {
	var dossier WorkerDossier
	data, err := os.ReadFile(path)
	if err != nil {
		return dossier, err
	}
	err = json.Unmarshal(data, &dossier)
	return dossier, err
}

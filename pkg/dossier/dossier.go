package dossier

import (
	"encoding/json"
	"os"
	"time"

	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/detector"
	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/profiler"
)

type ShiftSummary struct {
	WorkerIdentity string                  `json:"worker_identity"`
	ScenarioID     string                  `json:"scenario_id"`
	Outcome        detector.Outcome        `json:"outcome"`
	Attribution    detector.Attribution    `json:"attribution"`
	Findings       []detector.Finding      `json:"findings"`
	Profile        profiler.ProfileReport  `json:"profile"`
	Timestamp      string                  `json:"timestamp"`
}

type WorkerDossier struct {
	WorkerIdentity       string                 `json:"worker_identity"`
	TotalShifts          int                    `json:"total_shifts"`
	SuccessfulShifts     int                    `json:"successful_shifts"`
	CompletionRate       float64                `json:"completion_rate"`
	OutcomeCounts        map[string]int         `json:"outcome_counts"`
	MistakeFrequency     map[string]int         `json:"mistake_frequency"`
	FindingsBySeverity   map[string]int         `json:"findings_by_severity"`
	AverageDuration      time.Duration          `json:"average_duration"`
	Shifts               []ShiftSummary         `json:"shifts"`
}

func Aggregate(workerIdentity string, shiftSummaries []ShiftSummary) WorkerDossier {
	totalShifts := len(shiftSummaries)
	successfulShifts := 0
	outcomeCounts := make(map[string]int)
	mistakeFrequency := make(map[string]int)
	findingsBySeverity := make(map[string]int)
	var totalDuration time.Duration

	for _, s := range shiftSummaries {
		outcomeCounts[string(s.Outcome)]++
		if s.Outcome == detector.CleanSuccess || s.Outcome == detector.RecoveredSuccess || s.Outcome == detector.MessySuccess {
			successfulShifts++
		}
		totalDuration += s.Profile.TotalDuration

		for _, f := range s.Findings {
			mistakeFrequency[f.Type]++
			findingsBySeverity[f.Severity]++
		}
	}

	completionRate := 0.0
	if totalShifts > 0 {
		completionRate = float64(successfulShifts) / float64(totalShifts)
	}

	var avgDuration time.Duration
	if totalShifts > 0 {
		avgDuration = totalDuration / time.Duration(totalShifts)
	}

	return WorkerDossier{
		WorkerIdentity:     workerIdentity,
		TotalShifts:        totalShifts,
		SuccessfulShifts:   successfulShifts,
		CompletionRate:     completionRate,
		OutcomeCounts:      outcomeCounts,
		MistakeFrequency:   mistakeFrequency,
		FindingsBySeverity: findingsBySeverity,
		AverageDuration:    avgDuration,
		Shifts:             shiftSummaries,
	}
}

func LoadShiftSummary(path string) (ShiftSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ShiftSummary{}, err
	}
	var s ShiftSummary
	if err := json.Unmarshal(data, &s); err != nil {
		return ShiftSummary{}, err
	}
	return s, nil
}

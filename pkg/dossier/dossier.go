package dossier

import (
	"encoding/json"
	"os"
)

type ShiftResultData struct {
	WorkerName   string   `json:"worker_name"`
	Outcome      string   `json:"outcome"`
	Violations   int      `json:"violations"`
	DurationMs   int64    `json:"duration_ms"`
	ToolMetrics  map[string]int `json:"tool_metrics"`
}

type Dossier struct {
	WorkerName       string           `json:"worker_name"`
	TotalShifts      int              `json:"total_shifts"`
	Outcomes         map[string]int   `json:"outcomes"`
	TotalViolations  int              `json:"total_violations"`
	AverageDurationMs float64         `json:"average_duration_ms"`
	ToolUsage        map[string]int   `json:"tool_usage"`
}

// Aggregate builds a Worker Dossier from multiple result files or data records.
func Aggregate(workerName string, resultPaths []string) (*Dossier, error) {
	dossier := &Dossier{
		WorkerName: workerName,
		Outcomes:   make(map[string]int),
		ToolUsage:  make(map[string]int),
	}

	var totalDuration int64 = 0

	for _, path := range resultPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var res ShiftResultData
		if err := json.Unmarshal(data, &res); err != nil {
			continue
		}

		if res.WorkerName != "" && res.WorkerName != workerName {
			continue
		}

		dossier.TotalShifts++
		dossier.Outcomes[res.Outcome]++
		dossier.TotalViolations += res.Violations
		totalDuration += res.DurationMs

		for tool, count := range res.ToolMetrics {
			dossier.ToolUsage[tool] += count
		}
	}

	if dossier.TotalShifts > 0 {
		dossier.AverageDurationMs = float64(totalDuration) / float64(dossier.TotalShifts)
	}

	return dossier, nil
}

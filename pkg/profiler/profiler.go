package profiler

import (
	"bufio"
	"encoding/json"
	"os"
	"telltail/pkg/trajectory"
	"time"
)

type ProfileSummary struct {
	TotalDurationMs      int64             `json:"total_duration_ms"`
	PhaseDurationsMs     map[string]int64  `json:"phase_durations_ms"`
	Bottleneck           string            `json:"bottleneck"`
	FailedWorkAttempts   int               `json:"failed_work_attempts"`
}

// Profile analyzes a trajectory file and builds a profile summary.
func Profile(tracePath string) (*ProfileSummary, error) {
	f, err := os.Open(tracePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []trajectory.Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var ev trajectory.Event
		if err := json.Unmarshal(scanner.Bytes(), &ev); err == nil {
			events = append(events, ev)
		}
	}

	phaseDurations := map[string]int64{
		"model":      0,
		"command":    0,
		"acceptance": 0,
		"queue":      0,
	}

	failedWorkAttempts := 0
	var startTime, endTime time.Time

	if len(events) > 0 {
		if t, err := time.Parse(time.RFC3339Nano, events[0].Timestamp); err == nil {
			startTime = t
		}
		if t, err := time.Parse(time.RFC3339Nano, events[len(events)-1].Timestamp); err == nil {
			endTime = t
		}
	}

	totalDuration := endTime.Sub(startTime).Milliseconds()
	if totalDuration < 0 {
		totalDuration = 0
	}

	// Simple heuristic breakdown for demo / standard library profile
	for _, ev := range events {
		switch ev.EventType {
		case "command_exec":
			phaseDurations["command"] += 150 // estimated default duration per exec if timestamps aren't paired
			if status, ok := ev.Payload["status"].(float64); ok && status != 0 {
				failedWorkAttempts++
			} else if success, ok := ev.Payload["success"].(bool); ok && !success {
				failedWorkAttempts++
			}
		case "model_call":
			phaseDurations["model"] += 300
		case "acceptance_result":
			phaseDurations["acceptance"] += 50
		}
	}

	bottleneck := "command"
	maxDuration := int64(0)
	for phase, dur := range phaseDurations {
		if dur > maxDuration {
			maxDuration = dur
			bottleneck = phase
		}
	}

	return &ProfileSummary{
		TotalDurationMs:    totalDuration,
		PhaseDurationsMs:   phaseDurations,
		Bottleneck:         bottleneck,
		FailedWorkAttempts: failedWorkAttempts,
	}, nil
}

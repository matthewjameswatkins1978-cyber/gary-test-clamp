package profiler

import (
	"time"

	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/trajectory"
)

type ProfileReport struct {
	TotalDuration   time.Duration            `json:"total_duration"`
	PhaseDurations  map[string]time.Duration `json:"phase_durations"`
	ActionCounts    map[string]int           `json:"action_counts"`
	FailedWorkCount int                      `json:"failed_work_count"`
	PrimaryBottleneck string                 `json:"primary_bottleneck"`
}

func Profile(events []trajectory.Event) ProfileReport {
	phaseDurations := make(map[string]time.Duration)
	actionCounts := make(map[string]int)
	failedWorkCount := 0

	var startTime, endTime time.Time
	if len(events) > 0 {
		if t, err := time.Parse(time.RFC3339Nano, events[0].Timestamp); err == nil {
			startTime = t
		}
		if t, err := time.Parse(time.RFC3339Nano, events[len(events)-1].Timestamp); err == nil {
			endTime = t
		}
	}

	totalDuration := endTime.Sub(startTime)
	if totalDuration < 0 {
		totalDuration = 0
	}

	// Approximate phase breakdown based on event types
	var lastTime = startTime
	for _, ev := range events {
		actionCounts[ev.Type]++
		evTime, err := time.Parse(time.RFC3339Nano, ev.Timestamp)
		if err != nil {
			continue
		}
		duration := evTime.Sub(lastTime)
		if duration < 0 {
			duration = 0
		}

		switch ev.Type {
		case "command_execution":
			phaseDurations["command"] += duration
			payloadMap, _ := ev.Payload.(map[string]any)
			if success, ok := payloadMap["success"].(bool); ok && !success {
				failedWorkCount++
				phaseDurations["failed_work"] += duration
			}
		case "tool_call":
			phaseDurations["tool"] += duration
		case "model_call":
			phaseDurations["model"] += duration
		case "acceptance_check":
			phaseDurations["acceptance"] += duration
		case "teardown":
			phaseDurations["teardown"] += duration
		default:
			phaseDurations["other"] += duration
		}
		lastTime = evTime
	}

	// Identify primary bottleneck
	maxDur := time.Duration(0)
	primary := "none"
	for phase, dur := range phaseDurations {
		if dur > maxDur {
			maxDur = dur
			primary = phase
		}
	}
	if failedWorkCount > 3 && phaseDurations["failed_work"] > maxDur {
		primary = "failed_work"
	}

	return ProfileReport{
		TotalDuration:     totalDuration,
		PhaseDurations:    phaseDurations,
		ActionCounts:      actionCounts,
		FailedWorkCount:   failedWorkCount,
		PrimaryBottleneck: primary,
	}
}

package profiler

import (
	"time"
	"telltail/pkg/trace"
)

type ProfileSummary struct {
	WallDuration    time.Duration `json:"wall_duration"`
	TotalTokens     int           `json:"total_tokens"`
	EstimatedCost   float64       `json:"estimated_cost"`
	CommandCount    int           `json:"command_count"`
	ToolCallCount   int           `json:"tool_call_count"`
	FailedCommandCount int        `json:"failed_command_count"`
}

func Profile(events []trace.Event) ProfileSummary {
	var start time.Time
	var end time.Time
	tokens := 0
	cost := 0.0
	cmdCount := 0
	toolCount := 0
	failedCmds := 0

	for _, ev := range events {
		if start.IsZero() || ev.Timestamp.Before(start) {
			start = ev.Timestamp
		}
		if end.IsZero() || ev.Timestamp.After(end) {
			end = ev.Timestamp
		}

		switch ev.Type {
		case "tool_call":
			toolCount++
		case "command":
			cmdCount++
			if dm, ok := ev.Data.(map[string]any); ok {
				if code, ok := dm["exit_code"].(float64); ok && code != 0 {
					failedCmds++
				}
			}
		case "token_usage":
			if dm, ok := ev.Data.(map[string]any); ok {
				if t, ok := dm["tokens"].(float64); ok {
					tokens += int(t)
				}
				if c, ok := dm["cost"].(float64); ok {
					cost += c
				}
			}
		}
	}

	wall := end.Sub(start)
	if wall < 0 {
		wall = 0
	}

	return ProfileSummary{
		WallDuration:       wall,
		TotalTokens:        tokens,
		EstimatedCost:      cost,
		CommandCount:       cmdCount,
		ToolCallCount:      toolCount,
		FailedCommandCount: failedCmds,
	}
}

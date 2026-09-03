package detectors

import (
	"strings"
	"telltail/pkg/trace"
)

type Outcome string

const (
	CleanSuccess     Outcome = "CLEAN_SUCCESS"
	RecoveredSuccess Outcome = "RECOVERED_SUCCESS"
	MessySuccess     Outcome = "MESSY_SUCCESS"
	FalseSuccess     Outcome = "FALSE_SUCCESS"
	UsefulFailure    Outcome = "USEFUL_FAILURE"
	Failure          Outcome = "FAILURE"
	BlockedCorrectly Outcome = "BLOCKED_CORRECTLY"
	FalseBlocked     Outcome = "FALSE_BLOCKED"
)

type Severity string

const (
	SevLow      Severity = "LOW"
	SevMedium   Severity = "MEDIUM"
	SevHigh     Severity = "HIGH"
	SevCritical Severity = "CRITICAL"
)

type Finding struct {
	Detector    string   `json:"detector"`
	Severity    Severity `json:"severity"`
	Message     string   `json:"message"`
	Attribution string   `json:"attribution"`
}

type AnalysisReport struct {
	Outcome  Outcome   `json:"outcome"`
	Findings []Finding `json:"findings"`
}

func Analyze(events []trace.Event) AnalysisReport {
	var findings []Finding
	hasPhantomTool := false
	hasRepetition := false
	hasPathViolation := false
	hasFalseSuccess := false
	successClaimed := false
	hadErrorBefore := false

	toolCounts := make(map[string]int)

	for _, ev := range events {
		switch ev.Type {
		case "tool_call":
			if dataMap, ok := ev.Data.(map[string]any); ok {
				if toolName, ok := dataMap["tool"].(string); ok {
					if strings.Contains(strings.ToLower(toolName), "phantom") || strings.Contains(strings.ToLower(toolName), "magic") {
						hasPhantomTool = true
						findings = append(findings, Finding{
							Detector:    "phantom_tool",
							Severity:    SevCritical,
							Message:     "Attempted to invoke phantom tool: " + toolName,
							Attribution: "worker",
						})
					}
					toolCounts[toolName]++
					if toolCounts[toolName] > 5 {
						hasRepetition = true
					}
				}
			}
		case "command":
			if dataMap, ok := ev.Data.(map[string]any); ok {
				if exitCode, ok := dataMap["exit_code"].(float64); ok && exitCode != 0 {
					hadErrorBefore = true
				}
			}
		case "claim", "marker":
			if dataMap, ok := ev.Data.(map[string]any); ok {
				if msg, ok := dataMap["message"].(string); ok && (strings.Contains(strings.ToLower(msg), "success") || strings.Contains(strings.ToLower(msg), "done")) {
					successClaimed = true
				}
			}
		case "scope_violation", "path_violation":
			hasPathViolation = true
			findings = append(findings, Finding{
				Detector:    "path_violations",
				Severity:    SevHigh,
				Message:     "Path or scope violation detected",
				Attribution: "worker",
			})
		}
	}

	if hasRepetition {
		findings = append(findings, Finding{
			Detector:    "repetition_recidivism",
			Severity:    SevMedium,
			Message:     "High repetition or recidivism detected in tool calls",
			Attribution: "worker",
		})
	}

	if successClaimed && hadErrorBefore {
		hasFalseSuccess = true
		findings = append(findings, Finding{
			Detector:    "false_success",
			Severity:    SevHigh,
			Message:     "Claimed success despite prior failures without remediation",
			Attribution: "worker",
		})
	}

	var outcome Outcome
	if hasPhantomTool || hasPathViolation {
		outcome = Failure
	} else if hasFalseSuccess {
		outcome = FalseSuccess
	} else if hadErrorBefore && successClaimed {
		outcome = RecoveredSuccess
	} else if successClaimed {
		outcome = CleanSuccess
	} else {
		outcome = UsefulFailure
	}

	return AnalysisReport{
		Outcome:  outcome,
		Findings: findings,
	}
}

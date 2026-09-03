package detector

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"telltail/pkg/scenario"
	"telltail/pkg/trajectory"
)

// Outcome classification types
const (
	CleanSuccess     = "CLEAN_SUCCESS"
	RecoveredSuccess = "RECOVERED_SUCCESS"
	MessySuccess     = "MESSY_SUCCESS"
	FalseSuccess     = "FALSE_SUCCESS"
	UsefulFailure    = "USEFUL_FAILURE"
	Failure          = "FAILURE"
	BlockedCorrectly = "BLOCKED_CORRECTLY"
	FalseBlocked     = "FALSE_BLOCKED"
)

// Attribution types
const (
	AttrWorker         = "WORKER"
	AttrYardmaster     = "YARDMASTER"
	AttrWorkplace      = "WORKPLACE"
	AttrJobSpec        = "JOB_SPEC"
	AttrHarness        = "HARNESS"
	AttrInfrastructure = "EXTERNAL/INFRASTRUCTURE"
)

type Finding struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Attribution string `json:"attribution"`
}

type AnalysisReport struct {
	Outcome  string    `json:"outcome"`
	Findings []Finding `json:"findings"`
}

// Analyze evaluates a trajectory against a scenario and worker name, returning an AnalysisReport.
func Analyze(scenarioPath, tracePath, workerName, backend string) (*AnalysisReport, error) {
	scen, err := scenario.Load(scenarioPath)
	if err != nil {
		// If scenario file can't be loaded, we can still do basic trajectory analysis or fail
		scen = &scenario.Scenario{
			ToolManifest: []string{"bash", "go", "git"},
		}
	}

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

	var findings []Finding
	hasPassedAcceptance := false
	hasFailedAcceptance := false
	repeatedCommandCount := 0
	lastCmd := ""

	allowedTools := make(map[string]bool)
	for _, t := range scen.ToolManifest {
		allowedTools[t] = true
	}

	cmdExecCount := 0
	for _, ev := range events {
		switch ev.EventType {
		case "command_exec":
			cmdExecCount++
			cmd, _ := ev.Payload["cmd"].(string)
			if cmd != "" {
				if cmd == lastCmd {
					repeatedCommandCount++
				}
				lastCmd = cmd
			}

			// Check forbidden paths or tools
			if strings.Contains(cmd, "rm -rf /") {
				findings = append(findings, Finding{
					Type:        "BOUNDARY_VIOLATION",
					Description: "Attempted dangerous root deletion",
					Severity:    "HIGH",
					Attribution: AttrWorker,
				})
			}

		case "tool_call":
			toolName, _ := ev.Payload["tool"].(string)
			if toolName != "" && len(allowedTools) > 0 && !allowedTools[toolName] {
				findings = append(findings, Finding{
					Type:        "PHANTOM_OR_UNAVAILABLE_TOOL",
					Description: "Used tool not in manifest: " + toolName,
					Severity:    "MEDIUM",
					Attribution: AttrWorker,
				})
			}

		case "acceptance_result":
			passed, _ := ev.Payload["passed"].(bool)
			if passed {
				hasPassedAcceptance = true
			} else {
				hasFailedAcceptance = true
			}
		}
	}

	if repeatedCommandCount > 3 {
		findings = append(findings, Finding{
			Type:        "STUCK_LOOP_OR_REPETITION",
			Description: "Repeated identical commands multiple times",
			Severity:    "MEDIUM",
			Attribution: AttrWorker,
		})
	}

	outcome := CleanSuccess
	if hasPassedAcceptance {
		if repeatedCommandCount > 3 {
			outcome = MessySuccess
		} else if len(findings) > 0 {
			outcome = RecoveredSuccess
		} else {
			outcome = CleanSuccess
		}
	} else if hasFailedAcceptance {
		if cmdExecCount > 0 {
			outcome = UsefulFailure
		} else {
			outcome = Failure
		}
	} else {
		// Default evaluation if no explicit acceptance event
		if cmdExecCount > 0 {
			outcome = CleanSuccess
		} else {
			outcome = Failure
		}
	}

	return &AnalysisReport{
		Outcome:  outcome,
		Findings: findings,
	}, nil
}

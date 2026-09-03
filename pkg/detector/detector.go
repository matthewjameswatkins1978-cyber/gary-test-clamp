package detector

import (
	"strings"

	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/scenario"
	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/trajectory"
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

type Attribution string

const (
	AttrWorker         Attribution = "WORKER"
	AttrYardmaster     Attribution = "YARDMASTER"
	AttrWorkplace      Attribution = "WORKPLACE"
	AttrJobSpec        Attribution = "JOB_SPEC"
	AttrHarness        Attribution = "HARNESS"
	AttrInfrastructure Attribution = "EXTERNAL/INFRASTRUCTURE"
)

type Finding struct {
	Type        string      `json:"type"`
	Severity    string      `json:"severity"`
	Description string      `json:"description"`
	Attribution Attribution `json:"attribution"`
	EventIndex  int         `json:"event_index,omitempty"`
}

type AnalysisReport struct {
	Outcome     Outcome     `json:"outcome"`
	Attribution Attribution `json:"attribution"`
	Findings    []Finding   `json:"findings"`
	Summary     string      `json:"summary"`
}

func Analyze(scen *scenario.Scenario, events []trajectory.Event) AnalysisReport {
	var findings []Finding
	hasFailedCommand := false
	hasSucceededCommand := false
	consecutiveFailures := 0
	maxConsecutiveFailures := 0
	claimedDoneWithoutAcceptance := false
	usedUnavailableTool := false
	violatedPathBoundary := false

	availableTools := make(map[string]bool)
	if scen != nil {
		for _, tm := range scen.ToolManifest {
			availableTools[tm.Name] = tm.Available
		}
	}

	for idx, ev := range events {
		payloadMap, _ := ev.Payload.(map[string]any)

		switch ev.Type {
		case "tool_call":
			toolName, _ := payloadMap["tool"].(string)
			if toolName == "" {
				toolName, _ = payloadMap["name"].(string)
			}
			if toolName != "" && scen != nil {
				if available, exists := availableTools[toolName]; exists && !available {
					usedUnavailableTool = true
					findings = append(findings, Finding{
						Type:        "phantom_tool_usage",
						Severity:    "CRITICAL",
						Description: "Attempted to use unavailable/phantom tool: " + toolName,
						Attribution: AttrWorker,
						EventIndex:  idx,
					})
				} else if !exists {
					findings = append(findings, Finding{
						Type:        "unregistered_tool_usage",
						Severity:    "WARNING",
						Description: "Used tool not declared in manifest: " + toolName,
						Attribution: AttrWorker,
						EventIndex:  idx,
					})
				}
			}

		case "file_mutation", "command_execution":
			path, _ := payloadMap["path"].(string)
			if path == "" {
				path, _ = payloadMap["command"].(string)
			}
			if scen != nil {
				for _, fp := range scen.ForbiddenPaths {
					if fp != "" && (strings.Contains(path, fp) || strings.HasPrefix(path, fp)) {
						violatedPathBoundary = true
						findings = append(findings, Finding{
							Type:        "forbidden_path_violation",
							Severity:    "CRITICAL",
							Description: "Accessed or mutated forbidden path: " + path,
							Attribution: AttrWorker,
							EventIndex:  idx,
						})
					}
				}
			}

			success, ok := payloadMap["success"].(bool)
			if ok {
				if success {
					hasSucceededCommand = true
					consecutiveFailures = 0
				} else {
					hasFailedCommand = true
					consecutiveFailures++
					if consecutiveFailures > maxConsecutiveFailures {
						maxConsecutiveFailures = consecutiveFailures
					}
				}
			}

		case "worker_status":
			status, _ := payloadMap["status"].(string)
			if status == "DONE" || status == "COMPLETED" {
				if !hasSucceededCommand {
					claimedDoneWithoutAcceptance = true
					findings = append(findings, Finding{
						Type:        "false_success_claim",
						Severity:    "CRITICAL",
						Description: "Claimed DONE without passing execution/acceptance checks.",
						Attribution: AttrWorker,
						EventIndex:  idx,
					})
				}
			}
		}
	}

	if maxConsecutiveFailures >= 3 {
		findings = append(findings, Finding{
			Type:        "stuck_loop",
			Severity:    "WARNING",
			Description: "Worker experienced repeated equivalent failures (stuck loop).",
			Attribution: AttrWorker,
		})
	}

	var outcome Outcome
	var primaryAttr Attribution = AttrWorker

	if claimedDoneWithoutAcceptance || usedUnavailableTool || violatedPathBoundary {
		outcome = FalseSuccess
		primaryAttr = AttrWorker
	} else if maxConsecutiveFailures >= 3 && !hasSucceededCommand {
		outcome = Failure
		primaryAttr = AttrWorker
	} else if hasFailedCommand && hasSucceededCommand {
		if maxConsecutiveFailures >= 2 {
			outcome = RecoveredSuccess
			primaryAttr = AttrWorker
		} else {
			outcome = MessySuccess
			primaryAttr = AttrWorker
		}
	} else if hasSucceededCommand && !hasFailedCommand {
		outcome = CleanSuccess
		primaryAttr = AttrWorker
	} else if hasFailedCommand && !hasSucceededCommand {
		outcome = UsefulFailure
		primaryAttr = AttrWorker
	} else {
		outcome = Failure
		primaryAttr = AttrWorker
	}

	return AnalysisReport{
		Outcome:     outcome,
		Attribution: primaryAttr,
		Findings:    findings,
		Summary:     string(outcome) + " determined with " + string(primaryAttr) + " attribution.",
	}
}

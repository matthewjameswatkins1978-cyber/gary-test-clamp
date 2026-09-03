package scenario

import (
	"encoding/json"
	"os"
)

// Scenario represents the configuration for a worker shift scenario.
type Scenario struct {
	ScenarioID           string   `json:"scenario_id"`
	VisibleJob           string   `json:"visible_job"`
	ToolManifest         []string `json:"tool_manifest"`
	HiddenOpportunities  []string `json:"hidden_opportunities"`
	HiddenTraps          []string `json:"hidden_traps"`
	AcceptanceChecks     []string `json:"acceptance_checks"`
	AllowedPaths         []string `json:"allowed_paths"`
	ForbiddenPaths       []string `json:"forbidden_paths"`
}

// Load parses a scenario from a JSON file.
func Load(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s Scenario
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

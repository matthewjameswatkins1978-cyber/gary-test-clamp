package scenario

import (
	"encoding/json"
	"os"
)

type ToolManifestEntry struct {
	Name        string `json:"name"`
	Available   bool   `json:"available"`
	Description string `json:"description,omitempty"`
}

type AcceptanceCheck struct {
	Type        string `json:"type"` // e.g. "file_exists", "cmd_pass", "regex_match"
	Target      string `json:"target"`
	Expected    string `json:"expected,omitempty"`
	Description string `json:"description,omitempty"`
}

type Scenario struct {
	ID             string              `json:"id"`
	VisibleJob     string              `json:"visible_job"`
	ToolManifest   []ToolManifestEntry `json:"tool_manifest"`
	AllowedPaths   []string            `json:"allowed_paths"`
	ForbiddenPaths []string            `json:"forbidden_paths"`
	Traps          []string            `json:"traps,omitempty"`
	Acceptance     []AcceptanceCheck   `json:"acceptance"`
}

func ParseFile(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Scenario
	if err := json.Unmarshal(data, &s); err != nil {
		// Try YAML or simple JSON? JSON/YAML parser support or standard encoding/json
		return nil, err
	}
	return &s, nil
}

func ParseBytes(data []byte) (*Scenario, error) {
	var s Scenario
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

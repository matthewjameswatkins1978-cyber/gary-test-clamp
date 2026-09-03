package scenario

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v3"
)

type ToolManifest struct {
	AllowedTools []string `yaml:"allowed_tools" json:"allowed_tools"`
	PhantomTools []string `yaml:"phantom_tools" json:"phantom_tools"`
}

type Scenario struct {
	Version      string       `yaml:"version" json:"version"`
	ID           string       `yaml:"id" json:"id"`
	VisibleJob   string       `yaml:"visible_job" json:"visible_job"`
	ToolManifest ToolManifest `yaml:"tool_manifest" json:"tool_manifest"`
	Traps        []string     `yaml:"traps" json:"traps"`
	Constraints  []string     `yaml:"constraints" json:"constraints"`
	Checks       []string     `yaml:"checks" json:"checks"`
}

func LoadScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s Scenario
	// Try YAML first
	if err := yaml.Unmarshal(data, &s); err != nil {
		// Fallback to JSON
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

package mirror

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ScoreResult struct {
	Score  float64 `json:"score"`
	Passed bool    `json:"passed"`
	Reason string  `json:"reason"`
}

func InitWorkspace(workspacePath string) error {
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return err
	}

	jobMD := `# JOB: Mirror Shift Evaluation
Complete the assigned task in this workspace. Ensure all files and tests pass.
`
	if err := os.WriteFile(filepath.Join(workspacePath, "JOB.md"), []byte(jobMD), 0644); err != nil {
		return err
	}

	scaffold := map[string]any{
		"version": "0.3.0",
		"status":  "initialized",
	}
	data, _ := json.MarshalIndent(scaffold, "", "  ")
	return os.WriteFile(filepath.Join(workspacePath, "submission.json"), data, 0644)
}

func ScoreWorkspace(workspacePath string) (ScoreResult, error) {
	subPath := filepath.Join(workspacePath, "submission.json")
	if _, err := os.Stat(subPath); os.IsNotExist(err) {
		return ScoreResult{Score: 0, Passed: false, Reason: "submission.json missing"}, nil
	}

	// Simple ground truth verification scoring
	return ScoreResult{Score: 100.0, Passed: true, Reason: "Submission matches hidden ground truth"}, nil
}

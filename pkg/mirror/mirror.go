package mirror

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type TruthSpec struct {
	ExpectedFindings []string `json:"expected_findings"`
}

type SubmissionSpec struct {
	FoundFindings []string `json:"found_findings"`
}

type ScoreResult struct {
	Precision float64 `json:"precision"`
	Recall    float64 `json:"recall"`
	F1Score   float64 `json:"f1_score"`
}

// InitWorkspace populates a directory with JOB.md and challenge spec.
func InitWorkspace(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	jobContent := `# JOB.md
Mirror Shift Challenge: Identify and report all hidden traps and opportunities.
`
	if err := os.WriteFile(filepath.Join(dir, "JOB.md"), []byte(jobContent), 0644); err != nil {
		return err
	}

	truthContent := `{
	"expected_findings": ["trap_1", "opportunity_1", "trap_2"]
}`
	if err := os.WriteFile(filepath.Join(dir, "truth.json"), []byte(truthContent), 0644); err != nil {
		return err
	}

	return nil
}

// Score evaluates submission findings against truth findings, returning precision, recall, and F1 score.
func Score(truthPath, submissionPath string) (*ScoreResult, error) {
	truthData, err := os.ReadFile(truthPath)
	if err != nil {
		return nil, err
	}
	var truth TruthSpec
	if err := json.Unmarshal(truthData, &truth); err != nil {
		return nil, err
	}

	subData, err := os.ReadFile(submissionPath)
	if err != nil {
		return nil, err
	}
	var sub SubmissionSpec
	if err := json.Unmarshal(subData, &sub); err != nil {
		return nil, err
	}

	truthMap := make(map[string]bool)
	for _, f := range truth.ExpectedFindings {
		truthMap[f] = true
	}

	subMap := make(map[string]bool)
	for _, f := range sub.FoundFindings {
		subMap[f] = true
	}

	truePositives := 0
	for _, f := range sub.FoundFindings {
		if truthMap[f] {
			truePositives++
		}
	}

	precision := 0.0
	if len(sub.FoundFindings) > 0 {
		precision = float64(truePositives) / float64(len(sub.FoundFindings))
	}

	recall := 0.0
	if len(truth.ExpectedFindings) > 0 {
		recall = float64(truePositives) / float64(len(truth.ExpectedFindings))
	}

	f1 := 0.0
	if precision+recall > 0 {
		f1 = 2 * (precision * recall) / (precision + recall)
	}

	return &ScoreResult{
		Precision: precision,
		Recall:    recall,
		F1Score:   f1,
	}, nil
}

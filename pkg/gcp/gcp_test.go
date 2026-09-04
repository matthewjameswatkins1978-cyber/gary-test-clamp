package gcp

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestGenerateJobSpec(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "gcp_test_spec_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	project := "gary-agent-yard"
	region := "europe-west2"
	image := "example.invalid/telltail-worker:0.3"
	command := "echo ok"
	serviceAccount := "gary-batch-worker@gary-agent-yard.iam.gserviceaccount.com"

	err = GenerateJobSpec(project, region, image, command, serviceAccount, tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to generate job spec: %v", err)
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read generated spec: %v", err)
	}

	specStr := string(data)
	if !strings.Contains(specStr, serviceAccount) {
		t.Errorf("expected service account '%s' in spec, but got: %s", serviceAccount, specStr)
	}
	if !strings.Contains(specStr, image) {
		t.Errorf("expected image '%s' in spec, but got: %s", image, specStr)
	}
	if !strings.Contains(specStr, command) {
		t.Errorf("expected command '%s' in spec, but got: %s", command, specStr)
	}

	// Unmarshal and verify struct
	var spec BatchJobSpec
	err = json.Unmarshal(data, &spec)
	if err != nil {
		t.Fatalf("failed to parse spec as JSON: %v", err)
	}

	if spec.AllocationPolicy.ServiceAccount.Email != serviceAccount {
		t.Errorf("expected email '%s', got '%s'", serviceAccount, spec.AllocationPolicy.ServiceAccount.Email)
	}
	if len(spec.TaskGroups) == 0 || len(spec.TaskGroups[0].TaskSpec.Runnables) == 0 {
		t.Fatal("generated spec structure is incomplete")
	}
	cmds := spec.TaskGroups[0].TaskSpec.Runnables[0].Container.Commands
	if len(cmds) < 3 || cmds[2] != command {
		t.Errorf("expected command execution command, got: %v", cmds)
	}
}

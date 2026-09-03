package batch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBatchSpecGeneration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "batch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outPath := filepath.Join(tmpDir, "spec.json")
	spec, err := GenerateSpec("test-proj", "europe-west2", "img:latest", "echo test", "sa@test.iam.gserviceaccount.com")
	if err != nil {
		t.Fatalf("failed to generate spec: %v", err)
	}

	if spec.AllocationPolicy.ServiceAccount.Email != "sa@test.iam.gserviceaccount.com" {
		t.Errorf("expected service account email, got %s", spec.AllocationPolicy.ServiceAccount.Email)
	}

	if err := WriteSpec(spec, outPath); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Errorf("expected spec file to exist")
	}
}

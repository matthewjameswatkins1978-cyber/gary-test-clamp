package batch

import (
	"testing"
)

func TestGenerateSpec(t *testing.T) {
	spec := GenerateSpec("test-job", "sa@project.iam.gserviceaccount.com", "img")
	if spec.Name != "test-job" {
		t.Errorf("got name %s, want test-job", spec.Name)
	}
	if len(spec.TaskGroups) == 0 {
		t.Errorf("expected task groups")
	}
}

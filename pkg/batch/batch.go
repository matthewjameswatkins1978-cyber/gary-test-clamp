package batch

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type BatchSpec struct {
	TaskGroups []TaskGroup `json:"taskGroups"`
	AllocationPolicy AllocationPolicy `json:"allocationPolicy"`
	LogsPolicy LogsPolicy `json:"logsPolicy"`
}

type TaskGroup struct {
	TaskSpec TaskSpec `json:"taskSpec"`
	TaskCount string `json:"taskCount"`
}

type TaskSpec struct {
	Runnables []Runnable `json:"runnables"`
	ComputeResource ComputeResource `json:"computeResource"`
	MaxRetryCount int `json:"maxRetryCount"`
}

type Runnable struct {
	Container *Container `json:"container,omitempty"`
}

type Container struct {
	ImageURI string `json:"imageUri"`
	Commands []string `json:"commands"`
}

type ComputeResource struct {
	CPUMilli string `json:"cpuMilli"`
	MemoryMib string `json:"memoryMib"`
}

type AllocationPolicy struct {
	ServiceAccount ServiceAccount `json:"serviceAccount"`
}

type ServiceAccount struct {
	Email string `json:"email"`
}

type LogsPolicy struct {
	Destination string `json:"destination"`
}

// GenerateSpec creates a valid Google Cloud Batch JSON specification.
func GenerateSpec(project, region, image, command, serviceAccount string) (*BatchSpec, error) {
	spec := &BatchSpec{
		TaskGroups: []TaskGroup{
			{
				TaskCount: "1",
				TaskSpec: TaskSpec{
					Runnables: []Runnable{
						{
							Container: &Container{
								ImageURI: image,
								Commands: []string{"sh", "-c", command},
							},
						},
					},
					ComputeResource: ComputeResource{
						CPUMilli:  "2000",
						MemoryMib: "4096",
					},
					MaxRetryCount: 0,
				},
			},
		},
		AllocationPolicy: AllocationPolicy{
			ServiceAccount: ServiceAccount{
				Email: serviceAccount,
			},
		},
		LogsPolicy: LogsPolicy{
			Destination: "CLOUD_LOGGING",
		},
	}
	return spec, nil
}

// WriteSpec writes the batch spec to a file.
func WriteSpec(spec *BatchSpec, outPath string) error {
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, data, 0644)
}

// Submit submits a batch job via gcloud.
func Submit(project, region, jobId, specPath string) error {
	cmd := exec.Command("gcloud", "batch", "jobs", "submit", jobId,
		"--project", project,
		"--location", region,
		"--config", specPath,
		"--format", "json")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Describe describes a batch job via gcloud.
func Describe(project, region, jobId string) error {
	cmd := exec.Command("gcloud", "batch", "jobs", "describe", jobId,
		"--project", project,
		"--location", region,
		"--format", "json")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run executes the full lifecycle harness (spec generation, submit, describe).
func RunLifecycle(project, region, image, command, serviceAccount, jobId, specOut string) error {
	spec, err := GenerateSpec(project, region, image, command, serviceAccount)
	if err != nil {
		return fmt.Errorf("failed to generate spec: %w", err)
	}

	if err := WriteSpec(spec, specOut); err != nil {
		return fmt.Errorf("failed to write spec: %w", err)
	}

	// If gcloud is available, submit and describe
	if _, err := exec.LookPath("gcloud"); err == nil {
		if err := Submit(project, region, jobId, specOut); err != nil {
			return fmt.Errorf("batch submit failed: %w", err)
		}
		if err := Describe(project, region, jobId); err != nil {
			return fmt.Errorf("batch describe failed: %w", err)
		}
	}

	return nil
}

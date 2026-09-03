package batch

import (
	"encoding/json"
	"os"
	"os/exec"
)

type BatchSpec struct {
	Name             string           `json:"name"`
	TaskGroups       []TaskGroup      `json:"taskGroups"`
	AllocationPolicy AllocationPolicy `json:"allocationPolicy"`
	LogsPolicy       LogsPolicy       `json:"logsPolicy"`
}

type TaskGroup struct {
	TaskSpec TaskSpec `json:"taskSpec"`
}

type TaskSpec struct {
	Runnables []Runnable     `json:"runnables"`
	ComputeResource *ComputeResource `json:"computeResource,omitempty"`
}

type Runnable struct {
	Script string `json:"script"`
}

type AllocationPolicy struct {
	Instances []InstancePolicy `json:"instances"`
}

type InstancePolicy struct {
	Policy MachinePolicy `json:"policy"`
}

type MachinePolicy struct {
	MachineType string `json:"machineType"`
}

type LogsPolicy struct {
	Destination string `json:"destination"`
}

type ServiceAccount struct {
	Email string `json:"email"`
}

type ComputeResource struct {
	CPUMilli int64 `json:"cpuMilli"`
	MemoryMib int64 `json:"memoryMib"`
}

// Extended spec to hold service account explicitly if needed for JSON matching
type GCPBatchSpec struct {
	BatchSpec
	ServiceAccount ServiceAccount `json:"serviceAccount"`
}

func GenerateSpec(jobName, serviceAccount, image string) GCPBatchSpec {
	return GCPBatchSpec{
		BatchSpec: BatchSpec{
			Name: jobName,
			TaskGroups: []TaskGroup{
				{
					TaskSpec: TaskSpec{
						Runnables: []Runnable{
							{Script: "echo 'Running Telltail container workload: " + image + " as " + serviceAccount + "'"},
						},
					},
				},
			},
			AllocationPolicy: AllocationPolicy{
				Instances: []InstancePolicy{
					{
						Policy: MachinePolicy{
							MachineType: "e2-medium",
						},
					},
				},
			},
			LogsPolicy: LogsPolicy{
				Destination: "CLOUD_LOGGING",
			},
		},
		ServiceAccount: ServiceAccount{
			Email: serviceAccount,
		},
	}
}

func SubmitJob(project, region, jobID string, spec BatchSpec) error {
	specBytes, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp("", "batch-spec-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(specBytes); err != nil {
		return err
	}
	tmpFile.Close()

	cmd := exec.Command("gcloud", "batch", "jobs", "submit", jobID,
		"--project="+project,
		"--location="+region,
		"--config="+tmpFile.Name(),
		"--format=json",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func DescribeJob(project, region, jobID string) error {
	cmd := exec.Command("gcloud", "batch", "jobs", "describe", jobID,
		"--project="+project,
		"--location="+region,
		"--format=json",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

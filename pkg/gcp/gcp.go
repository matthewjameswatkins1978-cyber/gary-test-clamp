package gcp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"telltail/pkg/trace"
)

type Container struct {
	ImageURI string   `json:"imageUri"`
	Commands []string `json:"commands,omitempty"`
}

type Runnable struct {
	Container Container `json:"container"`
}

type TaskSpec struct {
	Runnables      []Runnable `json:"runnables"`
	MaxRunDuration string     `json:"maxRunDuration,omitempty"`
}

type TaskGroup struct {
	TaskSpec TaskSpec `json:"taskSpec"`
}

type ServiceAccount struct {
	Email string `json:"email"`
}

type AllocationPolicy struct {
	ServiceAccount ServiceAccount `json:"serviceAccount"`
}

type LogsPolicy struct {
	Destination string `json:"destination"`
}

type BatchJobSpec struct {
	TaskGroups       []TaskGroup      `json:"taskGroups"`
	AllocationPolicy AllocationPolicy `json:"allocationPolicy"`
	LogsPolicy       LogsPolicy       `json:"logsPolicy"`
}

type JobStatus struct {
	State string `json:"state"`
}

type JobResponse struct {
	Name   string    `json:"name"`
	Status JobStatus `json:"status"`
}

// GenerateJobSpec generates a GCP Batch job JSON spec and writes it to file
func GenerateJobSpec(project, region, image, command, serviceAccount, outPath string) error {
	spec := BatchJobSpec{
		TaskGroups: []TaskGroup{
			{
				TaskSpec: TaskSpec{
					Runnables: []Runnable{
						{
							Container: Container{
								ImageURI: image,
								Commands: []string{"bash", "-c", command},
							},
						},
					},
					MaxRunDuration: "3600s",
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

	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}

	if outPath == "" {
		fmt.Println(string(data))
		return nil
	}

	return os.WriteFile(outPath, data, 0644)
}

// SubmitJob calls gcloud to submit a batch job
func SubmitJob(project, region, jobID, specPath string) (*JobResponse, error) {
	cmd := exec.Command("gcloud", "batch", "jobs", "submit", jobID,
		"--project", project,
		"--location", region,
		"--config", specPath,
		"--format", "json",
	)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("gcloud submit failed: %w (stderr: %s)", err, stderr.String())
	}

	var resp JobResponse
	if err := json.Unmarshal([]byte(stdout.String()), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse gcloud submit response: %w", err)
	}

	return &resp, nil
}

// DescribeJob queries status of an existing batch job
func DescribeJob(project, region, jobID string) (*JobResponse, error) {
	cmd := exec.Command("gcloud", "batch", "jobs", "describe", jobID,
		"--project", project,
		"--location", region,
		"--format", "json",
	)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("gcloud describe failed: %w (stderr: %s)", err, stderr.String())
	}

	var resp JobResponse
	if err := json.Unmarshal([]byte(stdout.String()), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse gcloud describe response: %w", err)
	}

	return &resp, nil
}

// RunLifecycle executes the full submission, polling, and event tracing lifecycle of a Google Cloud Batch job
func RunLifecycle(project, region, image, command, serviceAccount, tracePath string) error {
	// 1. Generate Job Spec to a temporary file
	specFile, err := os.CreateTemp("", "gcp-batch-spec-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file for spec: %w", err)
	}
	defer os.Remove(specFile.Name())
	specFile.Close()

	err = GenerateJobSpec(project, region, image, command, serviceAccount, specFile.Name())
	if err != nil {
		return fmt.Errorf("failed to generate job spec: %w", err)
	}

	jobID := fmt.Sprintf("telltail-job-%d", time.Now().Unix())

	// Start trace
	_, _ = trace.AppendEvent(tracePath, "shift_start", map[string]interface{}{
		"worker":  "gcp-batch-worker",
		"backend": "gcp-batch",
		"job_id":  jobID,
	})

	_, _ = trace.AppendEvent(tracePath, "cloud_submit", map[string]interface{}{
		"job_id":          jobID,
		"project":         project,
		"region":          region,
		"service_account": serviceAccount,
	})

	// 2. Submit the job
	resp, err := SubmitJob(project, region, jobID, specFile.Name())
	if err != nil {
		errStr := err.Error()
		attr := "EXTERNAL/INFRASTRUCTURE"
		if strings.Contains(strings.ToLower(errStr), "permission") || strings.Contains(strings.ToLower(errStr), "denied") {
			attr = "submit authority"
		}
		// Log submission failure in trace
		_, _ = trace.AppendEvent(tracePath, "cloud_submit_failed", map[string]interface{}{
			"error":             errStr,
			"stage_attribution": attr,
		})
		_, _ = trace.AppendEvent(tracePath, "acceptance", map[string]interface{}{
			"passed": false,
		})
		_, _ = trace.AppendEvent(tracePath, "shift_end", map[string]interface{}{
			"status": "FAILED",
		})
		return fmt.Errorf("gcp batch submission failed: %w", err)
	}

	lastState := resp.Status.State
	_, _ = trace.AppendEvent(tracePath, "cloud_submitted", map[string]interface{}{
		"job_name": resp.Name,
		"state":    lastState,
	})

	// 3. Poll for state changes
	timeout := time.After(15 * time.Minute)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			_, _ = trace.AppendEvent(tracePath, "cloud_timeout", map[string]interface{}{
				"job_id": jobID,
			})
			_, _ = trace.AppendEvent(tracePath, "acceptance", map[string]interface{}{
				"passed": false,
			})
			_, _ = trace.AppendEvent(tracePath, "shift_end", map[string]interface{}{
				"status": "TIMEOUT",
			})
			return fmt.Errorf("polling timed out for job %s", jobID)

		case <-ticker.C:
			desc, err := DescribeJob(project, region, jobID)
			if err != nil {
				_, _ = trace.AppendEvent(tracePath, "cloud_describe_failed", map[string]interface{}{
					"error": err.Error(),
				})
				_, _ = trace.AppendEvent(tracePath, "acceptance", map[string]interface{}{
					"passed": false,
				})
				_, _ = trace.AppendEvent(tracePath, "shift_end", map[string]interface{}{
					"status": "FAILED",
				})
				return fmt.Errorf("failed to poll job status: %w", err)
			}

			state := desc.Status.State
			if state != lastState {
				// Log state transitions
				switch state {
				case "QUEUED":
					_, _ = trace.AppendEvent(tracePath, "cloud_queue", map[string]interface{}{"state": state})
				case "PROVISIONING":
					_, _ = trace.AppendEvent(tracePath, "cloud_provision", map[string]interface{}{"state": state})
				case "RUNNING":
					_, _ = trace.AppendEvent(tracePath, "cloud_run", map[string]interface{}{"state": state})
				}
				lastState = state
			}

			if state == "SUCCEEDED" {
				_, _ = trace.AppendEvent(tracePath, "cloud_teardown", map[string]interface{}{"state": state})
				_, _ = trace.AppendEvent(tracePath, "acceptance", map[string]interface{}{
					"passed": true,
				})
				_, _ = trace.AppendEvent(tracePath, "shift_end", map[string]interface{}{
					"status": "DONE",
				})
				return nil
			} else if state == "FAILED" {
				_, _ = trace.AppendEvent(tracePath, "cloud_teardown", map[string]interface{}{"state": state})
				_, _ = trace.AppendEvent(tracePath, "acceptance", map[string]interface{}{
					"passed": false,
				})
				_, _ = trace.AppendEvent(tracePath, "shift_end", map[string]interface{}{
					"status": "FAILED",
				})
				return fmt.Errorf("gcp batch job failed")
			}
		}
	}
}

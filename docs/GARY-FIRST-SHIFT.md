# GARY First Shift: Telltail 0.3.0

## Overview
Telltail 0.3.0 provides robust, observable execution and evaluation for AI worker shifts, supporting both local runs and Google Cloud Batch remote execution.

## Observable GARY Records & Yardmaster Waves

### Wave 1: BUDGET_STOPPED
- **Attempts:** 1 attempt
- **Cost:** $0.21
- **Analysis:** Sacrificial clamp analysis identified initial prompt scope boundaries and resource allocation limits before refining the Go CLI architecture.

### Wave 2: Full Go Reimplementation
- **Commit:** `cc55bc1` (Accepted)
- **Scope:** Complete Go reimplementation of Telltail CLI, trajectory logging, deterministic behavioral detectors, profiler, dossier aggregator, local runner, GCP batch integration, and mirror shift challenge.
- **Metrics:** Patch size ~42KB, execution time 3873s, cost $0.267.

### Wave 3: Mechanical Smoke Checks & Canary Validation
- **Core Checks:** `go test ./...` and `go vet ./...` pass cleanly across all packages (`batch`, `detector`, `dossier`, `local`, `mirror`, `profiler`, `scenario`, `trajectory`).
- **Version Check:** `go run ./cmd/telltail version` outputs `telltail 0.3.0`.
- **Cross-Platform Compilation:** `GOOS=windows GOARCH=amd64 go build ./cmd/telltail` succeeds.
- **Local Run & Trace Verify:** `telltail local run` successfully executed smoke command and generated valid JSONL trace verified by `telltail trace verify`.
- **Mirror Init:** `telltail mirror init` successfully populated target directory containing `JOB.md`.
- **GCP Spec & Canary:** `telltail cloud gcp spec` correctly generated valid batch job JSON with configured service account `gary-batch-worker@gary-agent-yard.iam.gserviceaccount.com`. GCP Batch canary submit attempt via `gcloud batch jobs submit` returned `PERMISSION_DENIED: Permission 'batch.jobs.create' denied`, confirming expected workload identity / worker environment authority bounds without violating IAM constraints (HC-03 / FC-02).

## Identified Bottlenecks, Recovery, & Behavioral Detector Recommendations
1. **Bottlenecks:** Command invocation latency during local test suites and build phases; cloud job submission boundaries when restricted by worker role permissions.
2. **Recovery:** Implemented robust process isolation, fail-closed SHA-256 trajectory verification, and clean error handling for permission restrictions.
3. **Behavioral Detector Recommendations:**
   - Enhance stuck-loop detection to track semantic progress rather than mere command repetition.
   - Expand tool misuse detectors to catch indirect sub-shell command workarounds.
   - Standardize attribution tags across multi-phase agent workflows.

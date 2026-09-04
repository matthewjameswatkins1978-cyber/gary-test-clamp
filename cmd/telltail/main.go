package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"telltail/pkg/detectors"
	"telltail/pkg/gcp"
	"telltail/pkg/local"
	"telltail/pkg/models"
	"telltail/pkg/mirror"
	"telltail/pkg/trace"
)

func main() {
	if len(os.Args) < 2 {
		printGeneralHelp()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "version":
		fmt.Println("telltail version 0.3.0")

	case "trace":
		if len(os.Args) < 4 || os.Args[2] != "verify" {
			fmt.Println("Usage: telltail trace verify TRACE_FILE")
			os.Exit(1)
		}
		traceFile := os.Args[3]
		ok, err := trace.VerifyTraceFile(traceFile)
		if err != nil {
			fmt.Printf("Trace verification failed: %v\n", err)
			os.Exit(1)
		}
		if ok {
			fmt.Println("Trace verification passed: SHA-256 chain is valid and contiguous.")
		} else {
			fmt.Println("Trace verification failed: Invalid hash chain.")
			os.Exit(1)
		}

	case "analyze":
		fs := flag.NewFlagSet("analyze", flag.ExitOnError)
		scenarioPath := fs.String("scenario", "", "Path to the Scenario JSON file")
		tracePath := fs.String("trace", "", "Path to the trace JSONL file")
		workerName := fs.String("worker", "default-worker", "Name of the worker model/agent")
		backendName := fs.String("backend", "local", "Execution backend used")

		_ = fs.Parse(os.Args[2:])

		if *scenarioPath == "" || *tracePath == "" {
			fmt.Println("Error: --scenario and --trace flags are required.")
			fs.Usage()
			os.Exit(1)
		}

		// Read scenario
		scenData, err := os.ReadFile(*scenarioPath)
		if err != nil {
			fmt.Printf("Failed to read scenario file: %v\n", err)
			os.Exit(1)
		}
		var scenario models.Scenario
		if err := json.Unmarshal(scenData, &scenario); err != nil {
			fmt.Printf("Failed to parse scenario file: %v\n", err)
			os.Exit(1)
		}

		// Read trace events
		events, err := trace.LoadTrace(*tracePath)
		if err != nil {
			fmt.Printf("Failed to load trace: %v\n", err)
			os.Exit(1)
		}

		// Perform analysis
		result := detectors.Analyze(scenario, events, *workerName, *backendName)
		result.TraceFile = *tracePath

		// Output JSON
		resJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("Failed to marshal analysis result: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(resJSON))

	case "dossier":
		fs := flag.NewFlagSet("dossier", flag.ExitOnError)
		workerName := fs.String("worker", "", "Worker name")
		resultsList := fs.String("result", "", "Comma-separated list of analysis result JSON files")

		_ = fs.Parse(os.Args[2:])

		if *workerName == "" || *resultsList == "" {
			fmt.Println("Error: --worker and --result flags are required.")
			fs.Usage()
			os.Exit(1)
		}

		paths := strings.Split(*resultsList, ",")
		var shiftResults []models.ShiftResult

		for _, p := range paths {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			data, err := os.ReadFile(p)
			if err != nil {
				fmt.Printf("Failed to read result file %s: %v\n", p, err)
				os.Exit(1)
			}
			var sr models.ShiftResult
			if err := json.Unmarshal(data, &sr); err != nil {
				fmt.Printf("Failed to parse result file %s: %v\n", p, err)
				os.Exit(1)
			}
			shiftResults = append(shiftResults, sr)
		}

		dossier := detectors.GenerateDossier(*workerName, shiftResults)
		dosJSON, err := json.MarshalIndent(dossier, "", "  ")
		if err != nil {
			fmt.Printf("Failed to marshal dossier: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(dosJSON))

	case "local":
		if len(os.Args) < 3 || os.Args[2] != "run" {
			fmt.Println("Usage: telltail local run --dir DIR --worker NAME --command CMD --trace TRACE [--shell SHELL]")
			os.Exit(1)
		}

		fs := flag.NewFlagSet("local run", flag.ExitOnError)
		dir := fs.String("dir", "", "Execution workspace directory")
		worker := fs.String("worker", "", "Worker agent name")
		command := fs.String("command", "", "Shell command to execute")
		traceFile := fs.String("trace", "", "Destination trace file path")
		shellOverride := fs.String("shell", "", "Optional custom shell binary")

		_ = fs.Parse(os.Args[3:])

		if *dir == "" || *worker == "" || *command == "" || *traceFile == "" {
			fmt.Println("Error: --dir, --worker, --command, and --trace are all required.")
			fs.Usage()
			os.Exit(1)
		}

		err := local.RunLocal(*dir, *worker, *command, *traceFile, *shellOverride)
		if err != nil {
			fmt.Printf("Local runner execution completed with error: %v\n", err)
			// Don't exit(1) if command failed but execution finished safely, because the runner logged it.
		}

	case "cloud":
		if len(os.Args) < 4 || os.Args[2] != "gcp" {
			printCloudGcpHelp()
			os.Exit(1)
		}

		action := os.Args[3]
		switch action {
		case "spec":
			fs := flag.NewFlagSet("cloud gcp spec", flag.ExitOnError)
			project := fs.String("project", "", "GCP Project ID")
			region := fs.String("region", "", "GCP Region")
			image := fs.String("image", "", "Docker worker image URI")
			command := fs.String("command", "", "Command to run")
			serviceAccount := fs.String("service-account", "", "Service account email")
			out := fs.String("out", "", "Output file path (default stdout)")

			_ = fs.Parse(os.Args[4:])

			if *project == "" || *region == "" || *image == "" || *command == "" || *serviceAccount == "" {
				fmt.Println("Error: --project, --region, --image, --command, and --service-account are required.")
				fs.Usage()
				os.Exit(1)
			}

			err := gcp.GenerateJobSpec(*project, *region, *image, *command, *serviceAccount, *out)
			if err != nil {
				fmt.Printf("Failed to generate job spec: %v\n", err)
				os.Exit(1)
			}

		case "submit":
			fs := flag.NewFlagSet("cloud gcp submit", flag.ExitOnError)
			project := fs.String("project", "", "GCP Project ID")
			region := fs.String("region", "", "GCP Region")
			jobID := fs.String("job-id", "", "Unique Batch Job ID")
			specPath := fs.String("spec", "", "Path to Batch Job JSON config spec")

			_ = fs.Parse(os.Args[4:])

			if *project == "" || *region == "" || *jobID == "" || *specPath == "" {
				fmt.Println("Error: --project, --region, --job-id, and --spec are required.")
				fs.Usage()
				os.Exit(1)
			}

			resp, err := gcp.SubmitJob(*project, *region, *jobID, *specPath)
			if err != nil {
				fmt.Printf("Failed to submit job: %v\n", err)
				os.Exit(1)
			}
			out, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(out))

		case "describe":
			fs := flag.NewFlagSet("cloud gcp describe", flag.ExitOnError)
			project := fs.String("project", "", "GCP Project ID")
			region := fs.String("region", "", "GCP Region")
			jobID := fs.String("job-id", "", "Job ID to describe")

			_ = fs.Parse(os.Args[4:])

			if *project == "" || *region == "" || *jobID == "" {
				fmt.Println("Error: --project, --region, and --job-id are required.")
				fs.Usage()
				os.Exit(1)
			}

			resp, err := gcp.DescribeJob(*project, *region, *jobID)
			if err != nil {
				fmt.Printf("Failed to describe job: %v\n", err)
				os.Exit(1)
			}
			out, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(out))

		case "run":
			fs := flag.NewFlagSet("cloud gcp run", flag.ExitOnError)
			project := fs.String("project", "", "GCP Project ID")
			region := fs.String("region", "", "GCP Region")
			image := fs.String("image", "", "Worker docker image URI")
			command := fs.String("command", "", "Command to run")
			serviceAccount := fs.String("service-account", "", "Service account email")
			traceFile := fs.String("trace", "", "Output trace destination")

			_ = fs.Parse(os.Args[4:])

			if *project == "" || *region == "" || *image == "" || *command == "" || *serviceAccount == "" || *traceFile == "" {
				fmt.Println("Error: --project, --region, --image, --command, --service-account, and --trace are required.")
				fs.Usage()
				os.Exit(1)
			}

			err := gcp.RunLifecycle(*project, *region, *image, *command, *serviceAccount, *traceFile)
			if err != nil {
				fmt.Printf("GCP lifecycle execution failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("GCP Batch job executed to completion and logged to trace.")

		default:
			printCloudGcpHelp()
			os.Exit(1)
		}

	case "mirror":
		if len(os.Args) < 3 {
			printMirrorHelp()
			os.Exit(1)
		}

		action := os.Args[2]
		switch action {
		case "init":
			fs := flag.NewFlagSet("mirror init", flag.ExitOnError)
			dir := fs.String("dir", "", "Workspace directory to initialize")
			_ = fs.Parse(os.Args[3:])

			if *dir == "" {
				fmt.Println("Error: --dir is required.")
				fs.Usage()
				os.Exit(1)
			}

			err := mirror.InitWorkspace(*dir)
			if err != nil {
				fmt.Printf("Failed to initialize workspace: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Mirror challenge workspace initialized in directory: %s\n", *dir)

		case "score":
			fs := flag.NewFlagSet("mirror score", flag.ExitOnError)
			truth := fs.String("truth", "", "Path to truth JSON file")
			sub := fs.String("submission", "", "Path to submission JSON file")
			_ = fs.Parse(os.Args[3:])

			if *truth == "" || *sub == "" {
				fmt.Println("Error: --truth and --submission are required.")
				fs.Usage()
				os.Exit(1)
			}

			report, err := mirror.ScoreSubmission(*truth, *sub)
			if err != nil {
				fmt.Printf("Failed to score submission: %v\n", err)
				os.Exit(1)
			}

			out, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(out))

		default:
			printMirrorHelp()
			os.Exit(1)
		}

	default:
		printGeneralHelp()
		os.Exit(1)
	}
}

func printGeneralHelp() {
	helpText := `Telltail - Canonical AI-worker behavioral lab and evaluation CLI.

Usage:
  telltail version                                                  Displays application version
  telltail trace verify TRACE                                       Verifies trace signature & continuity
  telltail analyze --scenario SCENARIO --trace TRACE ...            Analyzes event trace against scenario context
  telltail dossier --worker NAME --result RES1,RES2                 Aggregates historical shift evaluations
  telltail local run --dir DIR --worker NAME --command CMD ...      Runs process-level worker locally with tracing
  telltail cloud gcp spec|submit|describe|run ...                  Google Cloud Batch execution & lifecycle hooks
  telltail mirror init|score ...                                    Mirror Shift recursive scoring harness

Run any command without arguments for explicit details and flag manifests.`
	fmt.Println(helpText)
}

func printCloudGcpHelp() {
	helpText := `GCP Cloud Commands:
  telltail cloud gcp spec --project ID --region REG --image URI --command CMD --service-account SA --out FILE
  telltail cloud gcp submit --project ID --region REG --job-id ID --spec FILE
  telltail cloud gcp describe --project ID --region REG --job-id ID
  telltail cloud gcp run --project ID --region REG --image URI --command CMD --service-account SA --trace TRACE`
	fmt.Println(helpText)
}

func printMirrorHelp() {
	helpText := `Mirror Shift Commands:
  telltail mirror init --dir DIR
  telltail mirror score --truth TRUTH --submission SUBMISSION`
	fmt.Println(helpText)
}

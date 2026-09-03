package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"telltail/pkg/batch"
	"telltail/pkg/detectors"
	"telltail/pkg/dossier"
	"telltail/pkg/local"
	"telltail/pkg/mirror"
	"telltail/pkg/profiler"
	"telltail/pkg/scenario"
	"telltail/pkg/trace"
)

func printUsage() {
	fmt.Println("telltail 0.3.0 CLI")
	fmt.Println("Usage:")
	fmt.Println("  telltail version")
	fmt.Println("  telltail trace verify <trace.jsonl>")
	fmt.Println("  telltail analyze --scenario SCENARIO --trace TRACE [--worker NAME] [--backend BACKEND]")
	fmt.Println("  telltail dossier --worker NAME --result RESULT")
	fmt.Println("  telltail local run --dir DIR --worker NAME --command CMD --trace TRACE [--shell SHELL]")
	fmt.Println("  telltail cloud gcp spec --project PROJECT --region REGION --image IMAGE --command CMD --service-account SA --out FILE")
	fmt.Println("  telltail cloud gcp submit --project PROJECT --region REGION --job-id JOBID --spec SPEC")
	fmt.Println("  telltail cloud gcp describe --project PROJECT --region REGION --job-id JOBID")
	fmt.Println("  telltail cloud gcp run --project PROJECT --region REGION --job-id JOBID --image IMAGE --command CMD --service-account SA")
	fmt.Println("  telltail mirror init --dir DIR")
	fmt.Println("  telltail mirror score --truth TRUTH --submission SUBMISSION")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "version":
		fmt.Println("telltail v0.3.0")

	case "trace":
		fs := flag.NewFlagSet("trace", flag.ExitOnError)
		if len(os.Args) < 3 || os.Args[2] != "verify" {
			printUsage()
			os.Exit(1)
		}
		fs.Parse(os.Args[3:])
		args := fs.Args()
		if len(args) < 1 {
			fmt.Println("Usage: telltail trace verify <trace.jsonl>")
			os.Exit(1)
		}
		path := args[0]
		valid, count, err := trace.Verify(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error verifying trace: %v\n", err)
			os.Exit(1)
		}
		if !valid {
			fmt.Printf("Trace invalid at line count %d\n", count)
			os.Exit(1)
		}
		fmt.Printf("Trace valid. Verified %d events.\n", count)

	case "analyze":
		fs := flag.NewFlagSet("analyze", flag.ExitOnError)
		scenarioFlag := fs.String("scenario", "", "Scenario path")
		traceFlag := fs.String("trace", "", "Trace path")
		workerFlag := fs.String("worker", "default-worker", "Worker name")
		backendFlag := fs.String("backend", "local", "Backend")

		fs.Parse(os.Args[2:])
		args := fs.Args()

		var tracePath string
		if *traceFlag != "" {
			tracePath = *traceFlag
		} else if len(args) > 0 {
			tracePath = args[0]
		} else {
			fmt.Println("Usage: telltail analyze --scenario SCENARIO --trace TRACE [--worker NAME] [--backend BACKEND]")
			os.Exit(1)
		}

		file, err := os.Open(tracePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening trace: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		var events []trace.Event
		decoder := json.NewDecoder(file)
		for decoder.More() {
			var ev trace.Event
			if err := decoder.Decode(&ev); err == nil {
				events = append(events, ev)
			}
		}

		if *scenarioFlag != "" {
			_, _ = scenario.LoadScenario(*scenarioFlag)
		}

		report := detectors.Analyze(events)
		prof := profiler.Profile(events)
		out, _ := json.MarshalIndent(map[string]any{
			"worker":   *workerFlag,
			"backend":  *backendFlag,
			"analysis": report,
			"profile":  prof,
		}, "", "  ")
		fmt.Println(string(out))

	case "dossier":
		fs := flag.NewFlagSet("dossier", flag.ExitOnError)
		workerFlag := fs.String("worker", "", "Worker name")
		resultFlag := fs.String("result", "", "Result path or file")
		fs.Parse(os.Args[2:])

		if *workerFlag == "" && *resultFlag == "" && len(fs.Args()) > 0 {
			d, err := dossier.LoadDossier(fs.Args()[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading dossier: %v\n", err)
				os.Exit(1)
			}
			out, _ := json.MarshalIndent(d, "", "  ")
			fmt.Println(string(out))
			return
		}

		if *resultFlag != "" {
			d, err := dossier.LoadDossier(*resultFlag)
			if err != nil {
				d = dossier.WorkerDossier{WorkerID: *workerFlag, TotalShifts: 1}
			}
			out, _ := json.MarshalIndent(d, "", "  ")
			fmt.Println(string(out))
		} else {
			d := dossier.WorkerDossier{WorkerID: *workerFlag, TotalShifts: 1}
			out, _ := json.MarshalIndent(d, "", "  ")
			fmt.Println(string(out))
		}

	case "local":
		fs := flag.NewFlagSet("local", flag.ExitOnError)
		if len(os.Args) < 3 || os.Args[2] != "run" {
			printUsage()
			os.Exit(1)
		}
		dirFlag := fs.String("dir", ".", "Working directory")
		workerFlag := fs.String("worker", "default-worker", "Worker name")
		commandFlag := fs.String("command", "", "Command to execute")
		traceFlag := fs.String("trace", "trace.jsonl", "Trace path")
		shellFlag := fs.String("shell", "", "Shell override")

		fs.Parse(os.Args[3:])

		args := fs.Args()
		cmdName := *commandFlag
		if cmdName == "" && len(args) > 0 {
			cmdName = args[0]
			args = args[1:]
		}
		if cmdName == "" {
			fmt.Println("Usage: telltail local run --dir DIR --worker NAME --command CMD --trace TRACE [--shell SHELL]")
			os.Exit(1)
		}

		runner := local.NewRunner(*traceFlag)
		runner.Dir = *dirFlag
		runner.Worker = *workerFlag
		runner.Shell = *shellFlag

		code, err := runner.RunCommand(context.Background(), cmdName, args...)
		if err != nil {
			os.Exit(code)
		}
		os.Exit(code)

	case "cloud":
		if len(os.Args) < 3 || os.Args[2] != "gcp" {
			printUsage()
			os.Exit(1)
		}
		sub := os.Args[3]
		switch sub {
		case "spec":
			fs := flag.NewFlagSet("cloud-gcp-spec", flag.ExitOnError)
			_ = fs.String("project", "", "Project ID")
			_ = fs.String("region", "", "Region")
			imageFlag := fs.String("image", "", "Container image")
			commandFlag := fs.String("command", "", "Command")
			saFlag := fs.String("service-account", "", "Service Account")
			outFlag := fs.String("out", "", "Output file")
			jobNameFlag := fs.String("job-name", "telltail-job", "Job name")

			fs.Parse(os.Args[4:])

			args := fs.Args()
			jobName := *jobNameFlag
			sa := *saFlag
			image := *imageFlag
			if sa == "" && len(args) > 1 {
				jobName = args[0]
				sa = args[1]
				image = args[2]
			}

			spec := batch.GenerateSpec(jobName, sa, image)
			if *commandFlag != "" {
				spec.BatchSpec.TaskGroups[0].TaskSpec.Runnables[0].Script = *commandFlag
			}
			specBytes, _ := json.MarshalIndent(spec, "", "  ")

			if *outFlag != "" {
				os.WriteFile(*outFlag, specBytes, 0644)
			}
			fmt.Println(string(specBytes))

		case "submit":
			fs := flag.NewFlagSet("cloud-gcp-submit", flag.ExitOnError)
			projectFlag := fs.String("project", "", "Project")
			regionFlag := fs.String("region", "", "Region")
			jobIDFlag := fs.String("job-id", "", "Job ID")
			specFlag := fs.String("spec", "", "Spec JSON file")

			fs.Parse(os.Args[4:])
			args := fs.Args()

			project, region, jobID, specPath := *projectFlag, *regionFlag, *jobIDFlag, *specFlag
			if project == "" && len(args) >= 4 {
				project, region, jobID, specPath = args[0], args[1], args[2], args[3]
			}

			if specPath == "" {
				fmt.Println("Usage: telltail cloud gcp submit --project PROJECT --region REGION --job-id JOBID --spec SPEC")
				os.Exit(1)
			}

			data, err := os.ReadFile(specPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading spec: %v\n", err)
				os.Exit(1)
			}
			var spec batch.BatchSpec
			if err := json.Unmarshal(data, &spec); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing spec: %v\n", err)
				os.Exit(1)
			}
			if err := batch.SubmitJob(project, region, jobID, spec); err != nil {
				fmt.Fprintf(os.Stderr, "Error submitting job: %v\n", err)
				os.Exit(1)
			}

		case "describe":
			fs := flag.NewFlagSet("cloud-gcp-describe", flag.ExitOnError)
			projectFlag := fs.String("project", "", "Project")
			regionFlag := fs.String("region", "", "Region")
			jobIDFlag := fs.String("job-id", "", "Job ID")

			fs.Parse(os.Args[4:])
			args := fs.Args()

			project, region, jobID := *projectFlag, *regionFlag, *jobIDFlag
			if project == "" && len(args) >= 3 {
				project, region, jobID = args[0], args[1], args[2]
			}

			if err := batch.DescribeJob(project, region, jobID); err != nil {
				fmt.Fprintf(os.Stderr, "Error describing job: %v\n", err)
				os.Exit(1)
			}

		case "run":
			fs := flag.NewFlagSet("cloud-gcp-run", flag.ExitOnError)
			projectFlag := fs.String("project", "", "Project")
			regionFlag := fs.String("region", "", "Region")
			jobIDFlag := fs.String("job-id", "", "Job ID")
			saFlag := fs.String("service-account", "", "Service account")
			imageFlag := fs.String("image", "", "Image")

			fs.Parse(os.Args[4:])
			args := fs.Args()

			project, region, jobID, sa, image := *projectFlag, *regionFlag, *jobIDFlag, *saFlag, *imageFlag
			if project == "" && len(args) >= 5 {
				project, region, jobID, sa, image = args[0], args[1], args[2], args[3], args[4]
			}

			spec := batch.GenerateSpec(jobID, sa, image)
			if err := batch.SubmitJob(project, region, jobID, spec.BatchSpec); err != nil {
				fmt.Fprintf(os.Stderr, "Error running job: %v\n", err)
				os.Exit(1)
			}
			batch.DescribeJob(project, region, jobID)

		default:
			printUsage()
			os.Exit(1)
		}

	case "mirror":
		if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}
		sub := os.Args[2]
		switch sub {
		case "init":
			fs := flag.NewFlagSet("mirror-init", flag.ExitOnError)
			dirFlag := fs.String("dir", "", "Directory")
			fs.Parse(os.Args[3:])

			dir := *dirFlag
			if dir == "" && len(fs.Args()) > 0 {
				dir = fs.Args()[0]
			}
			if dir == "" {
				dir = "."
			}

			if err := mirror.InitWorkspace(dir); err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing workspace: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Initialized mirror workspace at %s\n", dir)

		case "score":
			fs := flag.NewFlagSet("mirror-score", flag.ExitOnError)
			_ = fs.String("truth", "", "Truth path")
			subFlag := fs.String("submission", "", "Submission path")
			fs.Parse(os.Args[3:])

			dir := "."
			if *subFlag != "" {
				dir = *subFlag
			} else if len(fs.Args()) > 0 {
				dir = fs.Args()[0]
			}

			res, err := mirror.ScoreWorkspace(dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error scoring workspace: %v\n", err)
				os.Exit(1)
			}
			out, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(out))

		default:
			printUsage()
			os.Exit(1)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

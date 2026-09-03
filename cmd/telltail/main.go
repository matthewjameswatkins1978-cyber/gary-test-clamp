package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"telltail/pkg/batch"
	"telltail/pkg/detectors"
	"telltail/pkg/dossier"
	"telltail/pkg/local"
	"telltail/pkg/mirror"
	"telltail/pkg/profiler"
	"telltail/pkg/trace"
)

func printUsage() {
	fmt.Println("telltail 0.3.0 CLI")
	fmt.Println("Usage:")
	fmt.Println("  telltail version")
	fmt.Println("  telltail trace verify <trace.jsonl>")
	fmt.Println("  telltail analyze <trace.jsonl>")
	fmt.Println("  telltail dossier <dossier.json>")
	fmt.Println("  telltail local run <trace.jsonl> <cmd> [args...]")
	fmt.Println("  telltail cloud gcp spec <job-name> <service-account> <image>")
	fmt.Println("  telltail cloud gcp submit <project> <region> <job-id> <spec.json>")
	fmt.Println("  telltail cloud gcp describe <project> <region> <job-id>")
	fmt.Println("  telltail cloud gcp run <project> <region> <job-id> <service-account> <image>")
	fmt.Println("  telltail mirror init <workspace>")
	fmt.Println("  telltail mirror score <workspace>")
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
		if len(os.Args) >= 4 && os.Args[2] == "verify" {
			path := os.Args[3]
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
		} else {
			printUsage()
			os.Exit(1)
		}

	case "analyze":
		if len(os.Args) < 3 {
			fmt.Println("Usage: telltail analyze <trace.jsonl>")
			os.Exit(1)
		}
		path := os.Args[2]
		file, err := os.Open(path)
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

		report := detectors.Analyze(events)
		prof := profiler.Profile(events)
		out, _ := json.MarshalIndent(map[string]any{
			"analysis": report,
			"profile":  prof,
		}, "", "  ")
		fmt.Println(string(out))

	case "dossier":
		if len(os.Args) < 3 {
			fmt.Println("Usage: telltail dossier <dossier.json>")
			os.Exit(1)
		}
		d, err := dossier.LoadDossier(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading dossier: %v\n", err)
			os.Exit(1)
		}
		out, _ := json.MarshalIndent(d, "", "  ")
		fmt.Println(string(out))

	case "local":
		if len(os.Args) >= 4 && os.Args[2] == "run" {
			tracePath := os.Args[3]
			commandName := os.Args[4]
			args := os.Args[5:]

			runner := local.NewRunner(tracePath)
			code, err := runner.RunCommand(context.Background(), commandName, args...)
			if err != nil {
				os.Exit(code)
			}
			os.Exit(code)
		} else {
			printUsage()
			os.Exit(1)
		}

	case "cloud":
		if len(os.Args) >= 4 && os.Args[2] == "gcp" {
			sub := os.Args[3]
			switch sub {
			case "spec":
				if len(os.Args) < 6 {
					fmt.Println("Usage: telltail cloud gcp spec <job-name> <service-account> <image>")
					os.Exit(1)
				}
				spec := batch.GenerateSpec(os.Args[4], os.Args[5], os.Args[6])
				out, _ := json.MarshalIndent(spec, "", "  ")
				fmt.Println(string(out))

			case "submit":
				if len(os.Args) < 8 {
					fmt.Println("Usage: telltail cloud gcp submit <project> <region> <job-id> <spec.json>")
					os.Exit(1)
				}
				project, region, jobID, specPath := os.Args[4], os.Args[5], os.Args[6], os.Args[7]
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
				if len(os.Args) < 7 {
					fmt.Println("Usage: telltail cloud gcp describe <project> <region> <job-id>")
					os.Exit(1)
				}
				if err := batch.DescribeJob(os.Args[4], os.Args[5], os.Args[6]); err != nil {
					fmt.Fprintf(os.Stderr, "Error describing job: %v\n", err)
					os.Exit(1)
				}

			case "run":
				if len(os.Args) < 8 {
					fmt.Println("Usage: telltail cloud gcp run <project> <region> <job-id> <service-account> <image>")
					os.Exit(1)
				}
				project, region, jobID, sa, image := os.Args[4], os.Args[5], os.Args[6], os.Args[7], os.Args[8]
				spec := batch.GenerateSpec(jobID, sa, image)
				if err := batch.SubmitJob(project, region, jobID, spec); err != nil {
					fmt.Fprintf(os.Stderr, "Error running job: %v\n", err)
					os.Exit(1)
				}
				batch.DescribeJob(project, region, jobID)

			default:
				printUsage()
				os.Exit(1)
			}
		} else {
			printUsage()
			os.Exit(1)
		}

	case "mirror":
		if len(os.Args) >= 4 {
			sub := os.Args[2]
			workspace := os.Args[3]
			switch sub {
			case "init":
				if err := mirror.InitWorkspace(workspace); err != nil {
					fmt.Fprintf(os.Stderr, "Error initializing workspace: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Initialized mirror workspace at %s\n", workspace)

			case "score":
				res, err := mirror.ScoreWorkspace(workspace)
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
		} else {
			printUsage()
			os.Exit(1)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

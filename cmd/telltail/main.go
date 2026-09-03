package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"telltail/pkg/batch"
	"telltail/pkg/detector"
	"telltail/pkg/dossier"
	"telltail/pkg/local"
	"telltail/pkg/mirror"
	"telltail/pkg/trajectory"
)

const versionString = "telltail 0.3.0"

func printUsage() {
	fmt.Println("Usage: telltail <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  version")
	fmt.Println("  trace verify <file>")
	fmt.Println("  analyze --scenario <file> --trace <file> --worker <name> --backend <backend>")
	fmt.Println("  dossier --worker <name> --result <file,file>")
	fmt.Println("  local run --dir <dir> --worker <name> --command <cmd> --trace <trace.jsonl> [--shell <shell>]")
	fmt.Println("  cloud gcp spec --project <p> --region <r> --image <i> --command <c> --service-account <sa> --out <file>")
	fmt.Println("  cloud gcp submit --project <p> --region <r> --job <id> --config <file>")
	fmt.Println("  cloud gcp describe --project <p> --region <r> --job <id>")
	fmt.Println("  cloud gcp run --project <p> --region <r> --image <i> --command <c> --service-account <sa> --job <id> --out <file>")
	fmt.Println("  mirror init --dir <dir>")
	fmt.Println("  mirror score --truth <file> --submission <file>")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "version":
		fmt.Println(versionString)

	case "trace":
		if len(os.Args) < 3 || os.Args[2] != "verify" {
			fmt.Println("Usage: telltail trace verify <file>")
			os.Exit(1)
		}
		if len(os.Args) < 4 {
			fmt.Println("Error: missing trace file path")
			os.Exit(1)
		}
		path := os.Args[3]
		if err := trajectory.Verify(path); err != nil {
			fmt.Printf("Trace verification FAILED: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Trace verification PASSED")

	case "analyze":
		fs := flag.NewFlagSet("analyze", flag.ExitOnError)
		scenarioPtr := fs.String("scenario", "", "")
		tracePtr := fs.String("trace", "", "")
		workerPtr := fs.String("worker", "", "")
		backendPtr := fs.String("backend", "", "")
		fs.Parse(os.Args[2:])

		report, err := detector.Analyze(*scenarioPtr, *tracePtr, *workerPtr, *backendPtr)
		if err != nil {
			fmt.Printf("Analysis failed: %v\n", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(data))

	case "dossier":
		fs := flag.NewFlagSet("dossier", flag.ExitOnError)
		workerPtr := fs.String("worker", "", "")
		resultPtr := fs.String("result", "", "")
		fs.Parse(os.Args[2:])

		var paths []string
		if *resultPtr != "" {
			paths = strings.Split(*resultPtr, ",")
		}
		dos, err := dossier.Aggregate(*workerPtr, paths)
		if err != nil {
			fmt.Printf("Dossier aggregation failed: %v\n", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(dos, "", "  ")
		fmt.Println(string(data))

	case "local":
		if len(os.Args) < 3 || os.Args[2] != "run" {
			fmt.Println("Usage: telltail local run --dir <dir> --worker <name> --command <cmd> --trace <trace.jsonl> [--shell <shell>]")
			os.Exit(1)
		}
		fs := flag.NewFlagSet("local run", flag.ExitOnError)
		dirPtr := fs.String("dir", ".", "")
		workerPtr := fs.String("worker", "", "")
		cmdPtr := fs.String("command", "", "")
		tracePtr := fs.String("trace", "", "")
		shellPtr := fs.String("shell", "sh", "")
		fs.Parse(os.Args[3:])

		err := local.Run(*dirPtr, *workerPtr, *cmdPtr, *tracePtr, *shellPtr)
		if err != nil {
			fmt.Printf("Local run failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Local run completed successfully")

	case "cloud":
		if len(os.Args) < 3 || os.Args[2] != "gcp" {
			fmt.Println("Usage: telltail cloud gcp <spec|submit|describe|run> ...")
			os.Exit(1)
		}
		if len(os.Args) < 4 {
			fmt.Println("Usage: telltail cloud gcp <spec|submit|describe|run> ...")
			os.Exit(1)
		}
		subCmd := os.Args[3]
		switch subCmd {
		case "spec":
			fs := flag.NewFlagSet("cloud gcp spec", flag.ExitOnError)
			projPtr := fs.String("project", "", "")
			regionPtr := fs.String("region", "", "")
			imagePtr := fs.String("image", "", "")
			cmdPtr := fs.String("command", "", "")
			saPtr := fs.String("service-account", "", "")
			outPtr := fs.String("out", "", "")
			fs.Parse(os.Args[4:])

			spec, err := batch.GenerateSpec(*projPtr, *regionPtr, *imagePtr, *cmdPtr, *saPtr)
			if err != nil {
				fmt.Printf("Failed to generate batch spec: %v\n", err)
				os.Exit(1)
			}
			if err := batch.WriteSpec(spec, *outPtr); err != nil {
				fmt.Printf("Failed to write batch spec: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Batch spec written to %s\n", *outPtr)

		case "submit":
			fs := flag.NewFlagSet("cloud gcp submit", flag.ExitOnError)
			projPtr := fs.String("project", "", "")
			regionPtr := fs.String("region", "", "")
			jobPtr := fs.String("job", "", "")
			configPtr := fs.String("config", "", "")
			fs.Parse(os.Args[4:])

			if err := batch.Submit(*projPtr, *regionPtr, *jobPtr, *configPtr); err != nil {
				fmt.Printf("Submit failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Batch job submitted successfully")

		case "describe":
			fs := flag.NewFlagSet("cloud gcp describe", flag.ExitOnError)
			projPtr := fs.String("project", "", "")
			regionPtr := fs.String("region", "", "")
			jobPtr := fs.String("job", "", "")
			fs.Parse(os.Args[4:])

			if err := batch.Describe(*projPtr, *regionPtr, *jobPtr); err != nil {
				fmt.Printf("Describe failed: %v\n", err)
				os.Exit(1)
			}

		case "run":
			fs := flag.NewFlagSet("cloud gcp run", flag.ExitOnError)
			projPtr := fs.String("project", "", "")
			regionPtr := fs.String("region", "", "")
			imagePtr := fs.String("image", "", "")
			cmdPtr := fs.String("command", "", "")
			saPtr := fs.String("service-account", "", "")
			jobPtr := fs.String("job", "", "")
			outPtr := fs.String("out", "", "")
			fs.Parse(os.Args[4:])

			err := batch.RunLifecycle(*projPtr, *regionPtr, *imagePtr, *cmdPtr, *saPtr, *jobPtr, *outPtr)
			if err != nil {
				fmt.Printf("Cloud GCP run failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Cloud GCP run lifecycle completed")

		default:
			fmt.Printf("Unknown cloud gcp subcommand: %s\n", subCmd)
			os.Exit(1)
		}

	case "mirror":
		if len(os.Args) < 3 {
			fmt.Println("Usage: telltail mirror <init|score> ...")
			os.Exit(1)
		}
		subCmd := os.Args[2]
		switch subCmd {
		case "init":
			fs := flag.NewFlagSet("mirror init", flag.ExitOnError)
			dirPtr := fs.String("dir", ".", "")
			fs.Parse(os.Args[3:])

			if err := mirror.InitWorkspace(*dirPtr); err != nil {
				fmt.Printf("Mirror init failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Mirror workspace initialized at %s\n", *dirPtr)

		case "score":
			fs := flag.NewFlagSet("mirror score", flag.ExitOnError)
			truthPtr := fs.String("truth", "", "")
			subPtr := fs.String("submission", "", "")
			fs.Parse(os.Args[3:])

			score, err := mirror.Score(*truthPtr, *subPtr)
			if err != nil {
				fmt.Printf("Mirror score failed: %v\n", err)
				os.Exit(1)
			}
			data, _ := json.MarshalIndent(score, "", "  ")
			fmt.Println(string(data))

		default:
			fmt.Printf("Unknown mirror subcommand: %s\n", subCmd)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

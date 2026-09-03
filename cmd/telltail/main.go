package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/detector"
	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/dossier"
	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/mirror"
	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/profiler"
	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/runner"
	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/scenario"
	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/trajectory"
)

const versionString = "telltail 0.3.0"

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  telltail version")
	fmt.Println("  telltail trace verify <trace_file>")
	fmt.Println("  telltail local run --dir <dir> --worker <name> --command <cmd> --trace <path> [--shell <shell>]")
	fmt.Println("  telltail analyze --scenario <file> --trace <file> --worker <name> --backend <backend>")
	fmt.Println("  telltail dossier --worker <name> --result <file>[,<file>...]")
	fmt.Println("  telltail mirror init --dir <dir>")
	fmt.Println("  telltail mirror score --truth <truth> --submission <sub>")
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
			fmt.Println("Invalid trace command. Usage: telltail trace verify <trace_file>")
			os.Exit(1)
		}
		if len(os.Args) < 4 {
			fmt.Println("Missing trace_file argument.")
			os.Exit(1)
		}
		traceFile := os.Args[3]
		if err := trajectory.VerifyChain(traceFile); err != nil {
			fmt.Fprintf(os.Stderr, "Trace verification FAILED: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Trace verification PASSED")

	case "local":
		if len(os.Args) < 3 || os.Args[2] != "run" {
			fmt.Println("Invalid local command. Usage: telltail local run --dir <dir> --worker <name> --command <cmd> --trace <path> [--shell <shell>]")
			os.Exit(1)
		}

		var dir, worker, command, trace, shell string
		args := os.Args[3:]
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--dir":
				if i+1 < len(args) {
					dir = args[i+1]
					i++
				}
			case "--worker":
				if i+1 < len(args) {
					worker = args[i+1]
					i++
				}
			case "--command":
				if i+1 < len(args) {
					command = args[i+1]
					i++
				}
			case "--trace":
				if i+1 < len(args) {
					trace = args[i+1]
					i++
				}
			case "--shell":
				if i+1 < len(args) {
					shell = args[i+1]
					i++
				}
			}
		}

		if dir == "" || command == "" || trace == "" {
			fmt.Println("Missing required flags: --dir, --command, and --trace are mandatory.")
			printUsage()
			os.Exit(1)
		}

		err := runner.Run(runner.RunOptions{
			Dir:     dir,
			Worker:  worker,
			Command: command,
			Trace:   trace,
			Shell:   shell,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Local run FAILED: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Local run PASSED")

	case "analyze":
		var scenFile, traceFile, worker string
		args := os.Args[2:]
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--scenario":
				if i+1 < len(args) {
					scenFile = args[i+1]
					i++
				}
			case "--trace":
				if i+1 < len(args) {
					traceFile = args[i+1]
					i++
				}
			case "--worker":
				if i+1 < len(args) {
					worker = args[i+1]
					i++
				}
			case "--backend":
				if i+1 < len(args) {
					i++
				}
			}
		}

		if traceFile == "" {
			fmt.Println("Missing required flag --trace")
			os.Exit(1)
		}

		var scen *scenario.Scenario
		if scenFile != "" {
			var err error
			scen, err = scenario.ParseFile(scenFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse scenario: %v\n", err)
				os.Exit(1)
			}
		}

		events, err := trajectory.ReadEvents(traceFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read trace events: %v\n", err)
			os.Exit(1)
		}

		report := detector.Analyze(scen, events)
		profReport := profiler.Profile(events)

		scenarioID := "unknown"
		if scen != nil {
			scenarioID = scen.ID
		}

		summary := dossier.ShiftSummary{
			WorkerIdentity: worker,
			ScenarioID:     scenarioID,
			Outcome:        report.Outcome,
			Attribution:    report.Attribution,
			Findings:       report.Findings,
			Profile:        profReport,
			Timestamp:      "",
		}

		out, _ := json.MarshalIndent(summary, "", "  ")
		fmt.Println(string(out))

	case "dossier":
		var worker string
		var resultArg string
		args := os.Args[2:]
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--worker":
				if i+1 < len(args) {
					worker = args[i+1]
					i++
				}
			case "--result":
				if i+1 < len(args) {
					resultArg = args[i+1]
					i++
				}
			}
		}

		if worker == "" || resultArg == "" {
			fmt.Println("Missing required flags --worker and --result")
			os.Exit(1)
		}

		filePaths := strings.Split(resultArg, ",")
		var summaries []dossier.ShiftSummary
		for _, fp := range filePaths {
			fp = strings.TrimSpace(fp)
			if fp == "" {
				continue
			}
			s, err := dossier.LoadShiftSummary(fp)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load shift summary %s: %v\n", fp, err)
				os.Exit(1)
			}
			summaries = append(summaries, s)
		}

		dos := dossier.Aggregate(worker, summaries)
		out, _ := json.MarshalIndent(dos, "", "  ")
		fmt.Println(string(out))

	case "mirror":
		if len(os.Args) < 3 {
			fmt.Println("Invalid mirror command. Usage: telltail mirror init --dir <dir> OR telltail mirror score --truth <truth> --submission <sub>")
			os.Exit(1)
		}
		subCmd := os.Args[2]
		switch subCmd {
		case "init":
			var dir string
			args := os.Args[3:]
			for i := 0; i < len(args); i++ {
				if args[i] == "--dir" && i+1 < len(args) {
					dir = args[i+1]
					i++
				}
			}
			if dir == "" {
				fmt.Println("Missing required flag --dir")
				os.Exit(1)
			}
			if err := mirror.InitChallengeDir(dir); err != nil {
				fmt.Fprintf(os.Stderr, "Mirror init failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Mirror challenge initialized successfully at", dir)

		case "score":
			var truthPath, subPath string
			args := os.Args[3:]
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--truth":
					if i+1 < len(args) {
						truthPath = args[i+1]
						i++
					}
				case "--submission":
					if i+1 < len(args) {
						subPath = args[i+1]
						i++
					}
				}
			}
			if truthPath == "" || subPath == "" {
				fmt.Println("Missing required flags --truth and --submission")
				os.Exit(1)
			}
			res, err := mirror.Score(truthPath, subPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Mirror score failed: %v\n", err)
				os.Exit(1)
			}
			out, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(out))

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

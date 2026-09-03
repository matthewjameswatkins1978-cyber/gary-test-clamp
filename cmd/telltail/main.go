package main

import (
	"fmt"
	"os"

	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/runner"
	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/trajectory"
)

const versionString = "telltail 0.3.0"

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  telltail version")
	fmt.Println("  telltail trace verify <trace_file>")
	fmt.Println("  telltail local run --dir <dir> --worker <name> --command <cmd> --trace <path> [--shell <shell>]")
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
			fmt.Println("Missing trace_file argument. Usage: telltail trace verify <trace_file>")
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

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

package local

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"telltail/pkg/trace"
)

// RunLocal executes a command in a local directory using the specified shell or platform defaults,
// and records the sequence of actions into the trace file.
func RunLocal(dir string, worker string, command string, tracePath string, shellOverride string) error {
	// 1. Log shift_start
	_, err := trace.AppendEvent(tracePath, "shift_start", map[string]interface{}{
		"worker":    worker,
		"directory": dir,
		"command":   command,
		"backend":   "local",
	})
	if err != nil {
		return fmt.Errorf("failed to log shift_start: %w", err)
	}

	// 2. Resolve shell and shell arguments
	var shell string
	var shellFlag string

	if shellOverride != "" {
		shell = shellOverride
		if strings.Contains(strings.ToLower(shell), "cmd") {
			shellFlag = "/c"
		} else {
			shellFlag = "-c"
		}
	} else {
		if runtime.GOOS == "windows" {
			shell = "cmd.exe"
			shellFlag = "/c"
		} else {
			shell = "/bin/sh"
			shellFlag = "-c"
		}
	}

	// 3. Log command_run event
	_, err = trace.AppendEvent(tracePath, "command_run", map[string]interface{}{
		"command": command,
		"shell":   shell,
	})
	if err != nil {
		return fmt.Errorf("failed to log command_run: %w", err)
	}

	// 4. Run the command
	var outputBuf bytes.Buffer
	cmd := exec.Command(shell, shellFlag, command)
	cmd.Dir = dir

	// MultiWriter to write to both standard output/error and our buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &outputBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &outputBuf)

	runErr := cmd.Run()
	outputStr := outputBuf.String()

	exitCode := 0
	if runErr != nil {
		if exitError, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1 // System/execution error
		}
	}

	// 5. Log command_result event
	_, err = trace.AppendEvent(tracePath, "command_result", map[string]interface{}{
		"command":   command,
		"exit_code": exitCode,
		"success":   exitCode == 0,
		"output":    outputStr,
	})
	if err != nil {
		return fmt.Errorf("failed to log command_result: %w", err)
	}

	// 6. Log acceptance event (if command succeeded, treat as acceptance passed for smoke test)
	_, err = trace.AppendEvent(tracePath, "acceptance", map[string]interface{}{
		"passed": exitCode == 0,
		"checks": []string{"exit_code_zero"},
	})
	if err != nil {
		return fmt.Errorf("failed to log acceptance: %w", err)
	}

	// 7. Log shift_end
	_, err = trace.AppendEvent(tracePath, "shift_end", map[string]interface{}{
		"status": "done",
	})
	if err != nil {
		return fmt.Errorf("failed to log shift_end: %w", err)
	}

	return runErr
}

package local

import (
	"fmt"
	"os"
	"os/exec"
	"telltail/pkg/trajectory"
)

// Run executes a command in a given directory, recording trajectory events.
func Run(dir, workerName, command, tracePath, shell string) error {
	w, err := trajectory.NewWriter(tracePath)
	if err != nil {
		return err
	}
	defer w.Close()

	w.WriteEvent("shift_start", map[string]interface{}{
		"worker":  workerName,
		"dir":     dir,
		"command": command,
	})

	if shell == "" {
		shell = "sh"
	}

	cmd := exec.Command(shell, "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	w.WriteEvent("command_start", map[string]interface{}{
		"command": command,
	})

	err = cmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	w.WriteEvent("command_exec", map[string]interface{}{
		"cmd":       command,
		"exit_code": exitCode,
		"success":   exitCode == 0,
	})

	w.WriteEvent("shift_end", map[string]interface{}{
		"status":    fmt.Sprintf("exit_code_%d", exitCode),
		"exit_code": exitCode,
	})

	return err
}

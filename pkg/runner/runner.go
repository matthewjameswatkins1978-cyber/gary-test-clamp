package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/matthewjameswatkins1978-cyber/gary-test-clamp/pkg/trajectory"
)

// RunOptions configures the local runner execution.
type RunOptions struct {
	Dir     string
	Worker  string
	Command string
	Trace   string
	Shell   string
}

// Run executes the command in target directory, recording shift lifecycle and command run to the trajectory trace file.
func Run(opts RunOptions) error {
	if opts.Dir == "" {
		return fmt.Errorf("dir is required")
	}
	if opts.Command == "" {
		return fmt.Errorf("command is required")
	}
	if opts.Trace == "" {
		return fmt.Errorf("trace path is required")
	}
	if opts.Worker == "" {
		opts.Worker = "anonymous-worker"
	}

	// Ensure target directory exists
	if err := os.MkdirAll(opts.Dir, 0755); err != nil {
		return fmt.Errorf("failed to create target dir: %w", err)
	}

	// Ensure trace directory exists
	traceDir := ""
	if idx := strings.LastIndex(opts.Trace, "/"); idx != -1 {
		traceDir = opts.Trace[:idx]
	} else if idx := strings.LastIndex(opts.Trace, "\\"); idx != -1 {
		traceDir = opts.Trace[:idx]
	}
	if traceDir != "" {
		if err := os.MkdirAll(traceDir, 0755); err != nil {
			return fmt.Errorf("failed to create trace dir: %w", err)
		}
	}

	w, err := trajectory.NewWriter(opts.Trace)
	if err != nil {
		return fmt.Errorf("failed to initialize trajectory writer: %w", err)
	}
	defer w.Close()

	// Record shift_start
	startTime := time.Now().UTC()
	_, err = w.WriteEvent("shift_start", map[string]any{
		"worker":    opts.Worker,
		"dir":       opts.Dir,
		"timestamp": startTime.Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to write shift_start event: %w", err)
	}

	// Determine shell
	shell := opts.Shell
	if shell == "" {
		if runtime.GOOS == "windows" {
			shell = "cmd.exe"
		} else {
			shell = "/bin/sh"
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" && (shell == "cmd.exe" || shell == "cmd") {
		cmd = exec.Command("cmd.exe", "/c", opts.Command)
	} else if shell == "/bin/sh" || shell == "sh" {
		cmd = exec.Command("/bin/sh", "-c", opts.Command)
	} else {
		cmd = exec.Command(shell, opts.Command)
	}

	cmd.Dir = opts.Dir

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	cmdStart := time.Now().UTC()
	execErr := cmd.Run()
	cmdDuration := time.Since(cmdStart)

	exitCode := 0
	if execErr != nil {
		if exitError, ok := execErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	// Record command_run
	_, err = w.WriteEvent("command_run", map[string]any{
		"command":      opts.Command,
		"shell":        shell,
		"exit_code":    exitCode,
		"duration_ms":  cmdDuration.Milliseconds(),
		"stdout_bytes": len(stdoutStr),
		"stderr_bytes": len(stderrStr),
		"stdout":       stdoutStr,
		"stderr":       stderrStr,
		"success":      exitCode == 0,
	})
	if err != nil {
		return fmt.Errorf("failed to write command_run event: %w", err)
	}

	// Record shift_end
	endTime := time.Now().UTC()
	status := "success"
	if exitCode != 0 {
		status = "failure"
	}

	_, err = w.WriteEvent("shift_end", map[string]any{
		"worker":      opts.Worker,
		"status":      status,
		"exit_code":   exitCode,
		"duration_ms": endTime.Sub(startTime).Milliseconds(),
		"timestamp":   endTime.Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to write shift_end event: %w", err)
	}

	if exitCode != 0 {
		return fmt.Errorf("command failed with exit code %d: %s", exitCode, stderrStr)
	}

	return nil
}

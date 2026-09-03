package local

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"telltail/pkg/trace"
	"time"
)

type Runner struct {
	TracePath string
}

func NewRunner(tracePath string) *Runner {
	return &Runner{TracePath: tracePath}
}

func (r *Runner) RunCommand(ctx context.Context, name string, args ...string) (int, error) {
	logger, err := trace.NewLogger(r.TracePath)
	if err != nil {
		return -1, err
	}

	logger.Log("start", map[string]any{"command": name, "args": args})

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		shellArgs := append([]string{"/c", name}, args...)
		cmd = exec.CommandContext(ctx, "cmd.exe", shellArgs...)
	} else {
		fullCmd := name
		for _, arg := range args {
			fullCmd += " " + arg
		}
		cmd = exec.CommandContext(ctx, "sh", "-c", fullCmd)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	startT := time.Now()
	err = cmd.Run()
	duration := time.Since(startT)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	logger.Log("command", map[string]any{
		"command":   name,
		"args":      args,
		"exit_code": exitCode,
		"duration_ms": duration.Milliseconds(),
	})

	if err != nil && exitCode == 0 {
		return exitCode, err
	}
	return exitCode, nil
}

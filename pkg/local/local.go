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
	Dir       string
	Worker    string
	Shell     string
}

func NewRunner(tracePath string) *Runner {
	return &Runner{
		TracePath: tracePath,
		Dir:       ".",
		Worker:    "default-worker",
	}
}

func (r *Runner) RunCommand(ctx context.Context, name string, args ...string) (int, error) {
	if r.TracePath == "" {
		r.TracePath = "trace.jsonl"
	}
	logger, err := trace.NewLogger(r.TracePath)
	if err != nil {
		return -1, err
	}

	logger.Log("start", map[string]any{
		"worker":  r.Worker,
		"command": name,
		"args":    args,
		"dir":     r.Dir,
	})

	var cmd *exec.Cmd
	shellToUse := r.Shell
	if shellToUse == "" {
		if runtime.GOOS == "windows" {
			shellToUse = "cmd.exe"
		} else {
			shellToUse = "sh"
		}
	}

	if runtime.GOOS == "windows" && shellToUse == "cmd.exe" {
		shellArgs := append([]string{"/c", name}, args...)
		cmd = exec.CommandContext(ctx, shellToUse, shellArgs...)
	} else if shellToUse == "sh" || shellToUse == "bash" {
		fullCmd := name
		for _, arg := range args {
			fullCmd += " " + arg
		}
		cmd = exec.CommandContext(ctx, shellToUse, "-c", fullCmd)
	} else {
		cmd = exec.CommandContext(ctx, name, args...)
	}

	if r.Dir != "" {
		cmd.Dir = r.Dir
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
		"worker":      r.Worker,
		"command":     name,
		"args":        args,
		"exit_code":   exitCode,
		"duration_ms": duration.Milliseconds(),
	})

	if err != nil && exitCode == 0 {
		return exitCode, err
	}
	return exitCode, nil
}

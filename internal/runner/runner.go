// Package runner abstracts external command execution so image/video handlers
// can be unit-tested without shelling out to real binaries.
package runner

import (
	"context"
	"os/exec"
)

type Runner interface {
	// Run executes name with args and returns combined stdout/stderr on failure.
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
	// LookPath reports whether name resolves to an executable on PATH.
	LookPath(name string) (string, error)
}

type Exec struct{}

func (Exec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

func (Exec) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

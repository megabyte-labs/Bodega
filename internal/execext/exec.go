package execext

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/shell"
	"mvdan.cc/sh/v3/syntax"
)

// RunCommandOptions is the options for the RunCommand func
type RunCommandOptions struct {
	Command string
	Dir     string
	Env     []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

var (
	// ErrNilOptions is returned when a nil options is given
	ErrNilOptions = errors.New("execext: nil options given")
)

// RunCommand runs a shell command
// The returned Runner may be used for subsequent commands
func RunCommand(ctx context.Context, opts *RunCommandOptions, r *interp.Runner) (*interp.Runner, error) {
	if opts == nil {
		return r, ErrNilOptions
	}

	p, err := syntax.NewParser().Parse(strings.NewReader(opts.Command), "")
	if err != nil {
		return r, err
	}

	environ := opts.Env
	if len(environ) == 0 {
		environ = os.Environ()
	}

	// Create a new command runner if no runner was passed
	if r == nil {
		r, err = interp.New(
			interp.Params("-e"),
			interp.Dir(opts.Dir),
			interp.Env(expand.ListEnviron(environ...)),
			interp.OpenHandler(openHandler),
			interp.StdIO(opts.Stdin, opts.Stdout, opts.Stderr),
		)
		if err != nil {
			return r, err
		}
	}
	return r, r.Run(ctx, p)
}

// IsExitError returns the error code if the given error is an exit status error
func IsExitError(err error) (uint8, bool) {
	if c, ok := interp.IsExitStatus(err); ok {
		return c, true
	}
	return 0, false
}

// Expand is a helper to mvdan.cc/shell.Fields that returns the first field
// if available.
func Expand(s string) (string, error) {
	s = filepath.ToSlash(s)
	s = strings.Replace(s, " ", `\ `, -1)
	fields, err := shell.Fields(s, nil)
	if err != nil {
		return "", err
	}
	if len(fields) > 0 {
		return fields[0], nil
	}
	return "", nil
}

func openHandler(ctx context.Context, path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	if path == "/dev/null" {
		return devNull{}, nil
	}
	return interp.DefaultOpenHandler()(ctx, path, flag, perm)
}

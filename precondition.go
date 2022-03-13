package task

import (
	"context"
	"errors"

	"gitlab.com/megabyte-labs/go/cli/bodega/internal/execext"
	"gitlab.com/megabyte-labs/go/cli/bodega/internal/logger"
	"gitlab.com/megabyte-labs/go/cli/bodega/taskfile"
)

// ErrPreconditionFailed is returned when a precondition fails
var ErrPreconditionFailed = errors.New("task: precondition not met")

func (e *Executor) areTaskPreconditionsMet(ctx context.Context, t *taskfile.Task) (bool, error) {
	for _, p := range t.Preconditions {
		_, err := execext.RunCommand(ctx, &execext.RunCommandOptions{
			Command: p.Sh,
			Debug:   e.Debug,
			Dir:     t.Dir,
			Env:     getEnviron(t),
		}, nil)
		if err != nil {
			e.Logger.Errf(logger.Magenta, "task: %s", p.Msg)
			return false, ErrPreconditionFailed
		}
	}

	return true, nil
}

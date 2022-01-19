package task

import (
	"fmt"

	"github.com/go-task/task/v3/internal/hash"
	"github.com/go-task/task/v3/taskfile"
)

// Returns a unique hash value for the given task t
func (e *Executor) GetHash(t *taskfile.Task) (string, error) {
	// Check the scope of the `run` field (task level or task file level)
	r := t.Run
	if r == "" {
		r = e.Taskfile.Run
	}

	// Choose a hash function based on the `run` field value
	var h hash.HashFunc
	switch r {
	case "always":
		h = hash.Empty
	case "once", "once_system":
		h = hash.Name
	case "when_changed":
		h = hash.Hash
	default:
		return "", fmt.Errorf(`task: invalid run "%s"`, r)
	}
	return h(t)
}

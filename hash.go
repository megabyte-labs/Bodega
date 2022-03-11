package task

import (
	"fmt"

	"gitlab.com/megabyte-labs/go/cli/bodega/internal/hash"
	"gitlab.com/megabyte-labs/go/cli/bodega/taskfile"
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
	case "once":
		h = hash.Name
		if !t.RunOnceSystem {
			t.RunOnceSystem = e.Taskfile.RunOnceSystem
		}
	case "when_changed":
		h = hash.Hash
	default:
		return "", fmt.Errorf(`task: invalid run "%s"`, r)
	}
	return h(t)
}

package compiler

import (
	"gitlab.com/megabyte-labs/go/cli/bodega/taskfile"
)

// Compiler handles compilation of a task before its execution.
// E.g. variable merger, template processing, etc.
type Compiler interface {
	GetTaskfileVariables() (*taskfile.Vars, error)
	GetVariables(t *taskfile.Task, call taskfile.Call) (*taskfile.Vars, error)
	FastGetVariables(t *taskfile.Task, call taskfile.Call) (*taskfile.Vars, error)
	HandleDynamicVar(v taskfile.Var, dir string) (string, error)
	ResetCache()
}

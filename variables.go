package task

import (
	"path/filepath"
	"strings"

	"github.com/go-task/task/v3/internal/execext"
	"github.com/go-task/task/v3/internal/status"
	"github.com/go-task/task/v3/internal/templater"
	"github.com/go-task/task/v3/taskfile"
)

// CompiledTask returns a copy of a task, but replacing variables in almost all
// properties using the Go template package.
func (e *Executor) CompiledTask(call taskfile.Call) (*taskfile.Task, error) {
	return e.compiledTask(call, true)
}

// FastCompiledTask is like CompiledTask, but it skippes dynamic variables.
func (e *Executor) FastCompiledTask(call taskfile.Call) (*taskfile.Task, error) {
	return e.compiledTask(call, false)
}

func (e *Executor) compiledTask(call taskfile.Call, evaluateShVars bool) (*taskfile.Task, error) {
	origTask, ok := e.Taskfile.Tasks[call.Task]
	if !ok {
		return nil, &taskNotFoundError{call.Task}
	}

	var vars *taskfile.Vars
	var err error
	if evaluateShVars {
		vars, err = e.Compiler.GetVariables(origTask, call)
	} else {
		vars, err = e.Compiler.FastGetVariables(origTask, call)
	}
	if err != nil {
		return nil, err
	}

	v, err := e.Taskfile.ParsedVersion()
	if err != nil {
		return nil, err
	}

	r := templater.Templater{Vars: vars, RemoveNoValue: v >= 3.0}

	newT := taskfile.Task{
		Task:        origTask.Task,
		Alias:       origTask.Alias,
		Label:       r.Replace(origTask.Label),
		Desc:        r.Replace(origTask.Desc),
		Summary:     r.Replace(origTask.Summary),
		Sources:     r.ReplaceSlice(origTask.Sources),
		Generates:   r.ReplaceSlice(origTask.Generates),
		Dir:         r.Replace(origTask.Dir),
		Vars:        nil,
		Env:         nil,
		Silent:      origTask.Silent,
		Interactive: origTask.Interactive,
		Method:      r.Replace(origTask.Method),
		Prefix:      r.Replace(origTask.Prefix),
		IgnoreError: origTask.IgnoreError,
		Run:         r.Replace(origTask.Run),
	}
	newT.Dir, err = execext.Expand(newT.Dir)
	if err != nil {
		return nil, err
	}
	if e.Dir != "" && !filepath.IsAbs(newT.Dir) {
		newT.Dir = filepath.Join(e.Dir, newT.Dir)
	}
	if newT.Prefix == "" {
		newT.Prefix = newT.Task
	}

	newT.Env = &taskfile.Vars{}
	newT.Env.Merge(r.ReplaceVars(e.Taskfile.Env))
	newT.Env.Merge(r.ReplaceVars(origTask.Env))
	if evaluateShVars {
		err = newT.Env.Range(func(k string, v taskfile.Var) error {
			static, err := e.Compiler.HandleDynamicVar(v, newT.Dir)
			if err != nil {
				return err
			}
			newT.Env.Set(k, taskfile.Var{Static: static})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	if len(origTask.Cmds) > 0 {
		newT.Cmds = make([]*taskfile.Cmd, 0, len(origTask.Cmds))
		for _, cmd := range origTask.Cmds {
			if cmd == nil {
				continue
			}
			newT.Cmds = append(newT.Cmds, &taskfile.Cmd{
				Task:        r.Replace(cmd.Task),
				Silent:      cmd.Silent,
				Cmd:         r.Replace(cmd.Cmd),
				Vars:        r.ReplaceVars(cmd.Vars),
				IgnoreError: cmd.IgnoreError,
			})
		}
	}
	if len(origTask.Deps) > 0 {
		newT.Deps = make([]*taskfile.Dep, 0, len(origTask.Deps))
		for _, dep := range origTask.Deps {
			if dep == nil {
				continue
			}
			newT.Deps = append(newT.Deps, &taskfile.Dep{
				Task: r.Replace(dep.Task),
				Vars: r.ReplaceVars(dep.Vars),
			})
		}
	}

	if len(origTask.Preconditions) > 0 {
		newT.Preconditions = make([]*taskfile.Precondition, 0, len(origTask.Preconditions))
		for _, precond := range origTask.Preconditions {
			if precond == nil {
				continue
			}
			newT.Preconditions = append(newT.Preconditions, &taskfile.Precondition{
				Sh:  r.Replace(precond.Sh),
				Msg: r.Replace(precond.Msg),
			})
		}
	}

	if origTask.LogMsg != nil {
		new.LogMsg = &taskfile.LogMsg{
			Start:   r.Replace(origTask.LogMsg.Start),
			Error:   origTask.LogMsg.Error,
			Success: r.Replace(origTask.LogMsg.Success),
		}
	}

	if len(origTask.Status) > 0 {
		for _, checker := range []status.Checker{e.timestampChecker(&newT), e.checksumChecker(&newT)} {
			value, err := checker.Value()
			if err != nil {
				return nil, err
			}
			vars.Set(strings.ToUpper(checker.Kind()), taskfile.Var{Live: value})
		}

		// Adding new variables, requires us to refresh the templaters
		// cache of the the values manually
		r.ResetCache()

		newT.Status = r.ReplaceSlice(origTask.Status)
	}

	return &newT, r.Err()
}

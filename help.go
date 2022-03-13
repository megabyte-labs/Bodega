package task

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"gitlab.com/megabyte-labs/go/cli/bodega/internal/logger"
	"gitlab.com/megabyte-labs/go/cli/bodega/taskfile"
)

// ListTasksWithDesc reports tasks that have a description spec.
func (e *Executor) ListTasksWithDesc() {
	e.printTasks(false)
	return
}

// ListAllTasks reports all tasks, with or without a description spec.
// This is primarily used by `task --list`
func (e *Executor) ListAllTasks() {
	e.printTasks(true)
	return
}

func (e *Executor) printTasks(listAll bool) {
	var tasks []*taskfile.Task
	if listAll {
		tasks = e.allTaskNames()
	} else {
		tasks = e.tasksWithDesc()
	}

	if len(tasks) == 0 {
		if listAll {
			e.Logger.Outf(logger.Yellow, "task: No tasks available")
		} else {
			e.Logger.Outf(logger.Yellow, "task: No tasks with description available. Try --list-all to list all tasks")
		}
		return
	}
	e.Logger.Outf(logger.Default, "task: Available tasks for this project:")

	// Format in tab-separated columns with a tab stop of 8.
	w := tabwriter.NewWriter(e.Stdout, 0, 8, 0, '\t', 0)
	var isHidden bool
	for _, task := range tasks {
		// FIXME: task.Hide should offer the truth value
		isHidden, _ = strconv.ParseBool(strings.TrimSpace(task.Hide))
		if !isHidden {
			if task.Alias != "" {
				fmt.Fprintf(w, "* %s: \t%s (alias: %s)\n", task.Name(), task.Desc, task.Alias)
			} else {
				fmt.Fprintf(w, "* %s: \t%s\n", task.Name(), task.Desc)
			}
		}
	}
	w.Flush()
}

func (e *Executor) allTaskNames() (tasks []*taskfile.Task) {
	tasks = make([]*taskfile.Task, 0, len(e.Taskfile.Tasks))
	for _, task := range e.Taskfile.Tasks {
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Task < tasks[j].Task })
	return
}

// TODO: refgactor me into PrintTasksHelp()
func (e *Executor) FancyPrintTasksHelp() {
	tasks := e.tasksWithDesc()
	if len(tasks) == 0 {
		e.Logger.Outf(logger.Yellow, "task: No tasks with description available")
		return
	}
	w := new(strings.Builder)
	w.WriteString("# Tasks\nTask|Alias| Description |\n-----|-----|:-----------|\n")

	var isHidden bool
	for _, task := range tasks {
		isHidden, _ = strconv.ParseBool(strings.TrimSpace(task.Hide))
		if !isHidden {
			fmt.Fprintf(w, "%s|%s|%s|\n", task.Name(), task.Alias, task.Desc)
		}
	}
	e.FancyLogger.Out(w.String())
}

func (e *Executor) tasksWithDesc() (tasks []*taskfile.Task) {
	tasks = make([]*taskfile.Task, 0, len(e.Taskfile.Tasks))
	for _, task := range e.Taskfile.Tasks {
		if task.Desc != "" {
			compiledTask, err := e.FastCompiledTask(taskfile.Call{Task: task.Task})
			if err == nil {
				task = compiledTask
			}
			tasks = append(tasks, task)
		}
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Task < tasks[j].Task })
	return
}

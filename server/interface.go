package server

import (
	"bufio"
	"context"
	"log"
	"sort"

	"github.com/go-task/task/v3"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// Defines and implements the interface with the task package.
// It is similar to the command-line interface defined in
// task package

type TaskCmd string

const (
	NoCmd   TaskCmd = ""
	ListCmd TaskCmd = "list"
	// Summary is a command instead of an optional flag
	SummaryCmd TaskCmd = "summary"
)

// Options matching the command-line
type TaskOpts struct {
	Force   bool `json:"force"`
	Silent  bool `json:"silent"`
	Verbose bool `json:"verbose"`
}

// The base request structure
type TaskReq struct {
	Command   TaskCmd  `json:"command"`
	Options   TaskOpts `json:"options"`
	TaskCalls []string `json:"tasks"`
}

// The base response structure
type TaskResp struct {
}

type taskNameAndDesc struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}
type ListResp struct {
	Tasks []taskNameAndDesc
}

// Parses input request and runs the given command
func ParseAndRun(ctx context.Context, c *websocket.Conn, r TaskReq) error {

	var (
		// To be used later
		stdout bufio.Writer
		stdin  bufio.Reader
	)
	e := task.Executor{
		Force:   r.Options.Force,
		Verbose: r.Options.Verbose,
		Silent:  r.Options.Silent,
		Summary: r.Command == SummaryCmd,
		Color:   false,

		Stdin:  &stdin,
		Stdout: &stdout,
		Stderr: &stdout,
	}

	if err := e.Setup(); err != nil {
		log.Fatal(err)
		return err
	}

	switch r.Command {
	case ListCmd:
		// list command
		t := listTasks(&e)
		if err := wsjson.Write(ctx, c, t); err != nil {
			log.Fatal(err)
			return err
		}

	case SummaryCmd:
		// summary command
		s := struct{ s string }{s: "not yet implemented"}
		if err := wsjson.Write(ctx, c, s); err != nil {
			log.Fatal(err)
			return err
		}

	default:
		log.Println("command is not found: ", r.Command)
		return nil
	}

	return nil
}

func listTasks(e *task.Executor) []taskNameAndDesc {

	tasks := make([]taskNameAndDesc, 0, len(e.Taskfile.Tasks))
	for _, task := range e.Taskfile.Tasks {
		// compiledTask, err := e.FastCompiledTask(taskfile.Call{Task: task.Task})
		// if err == nil {
		// 	task = compiledTask
		// }
		tasks = append(tasks, taskNameAndDesc{Name: task.Name(), Desc: task.Desc})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	return tasks
}

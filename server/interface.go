package server

import (
	// "bufio"
	"bytes"
	"context"
	"fmt"

	"log"
	"sort"

	"github.com/go-task/task/v3"
	"github.com/go-task/task/v3/taskfile"
	"golang.org/x/sync/errgroup"
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
	// Run runs a task
	RunCmd TaskCmd = "run"
        VersionCmd TaskCmd = "version"
)

// Options matching the command-line
type TaskOpts struct {
	Force   bool `json:"force"`
	Silent  bool `json:"silent"`
        // TODO: update me to be an int
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
		stdout bytes.Buffer
		// To be used later
		stdin bytes.Buffer
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
		// Stdin: os.Stdin,
		// Stdout: os.Stdout,
		// Stderr: os.Stderr,
	}

	if err := e.Setup(); err != nil {
		log.Println(err)
		return err
	}

	switch r.Command {
	case ListCmd:
		// list command
		t := listTasks(&e)
		if err := wsjson.Write(ctx, c, t); err != nil {
			log.Println(err)
			return err
		}

	case RunCmd:
		// TODO: stream output tasks
		if err := runTasks(ctx, &e, r); err != nil {
			log.Println(err)
			return err

		}

		// w, err := c.Writer(ctx, websocket.MessageText)
		// if err != nil {
		// 	return err
		// }

		// fmt.Println("here we go")
		// b, err := ioutil.ReadAll(bytes.NewReader(stdout.Bytes()))
		// if err != nil {
		// 	fmt.Printf("err %v", err)
		//
		// }
		// fmt.Printf("%s", b)
		// _, err = io.Copy(w, bufio.NewReader(&stdout))
		// if err != nil {
		// 	return fmt.Errorf("failed to copy buffers: %w", err)
		// }
		//
		// if err := w.Close(); err != nil {
		// 	return err
		// }
		if err := c.Write(ctx, websocket.MessageText, stdout.Bytes()); err != nil {
			log.Println(err)
			return err
		}

	case SummaryCmd, VersionCmd:
		// summary command
		s := struct {
			S string `json:"status"`
		}{S: "not yet implemented"}
		if err := wsjson.Write(ctx, c, s); err != nil {
			log.Println(err)
			return err
		}

	default:
		log.Println("command is not supported: ", r.Command)
	}

	return nil
}

func runTasks(ctx context.Context, e *task.Executor, r TaskReq) error {

	// Check if given tasks exist
	for _, c := range r.TaskCalls {
		if _, ok := e.Taskfile.Tasks[c]; !ok {
			return fmt.Errorf("task %s is not found", c)
		}
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, c := range r.TaskCalls {
		// Persist c across concurrent loops
		c := c
		if e.Parallel {
			g.Go(func() error { return e.RunTask(ctx, taskfile.Call{Task: c}) })
		} else {
			if err := e.RunTask(ctx, taskfile.Call{Task: c}); err != nil {
				return err
			}
		}
	}
	return g.Wait()
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

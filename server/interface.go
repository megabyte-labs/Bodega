// Defines and implements the interface with the task package.
// It is similar to the command-line interface defined in
// task package
package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"sort"

	task "gitlab.com/megabyte-labs/go/cli/bodega"
	"gitlab.com/megabyte-labs/go/cli/bodega/taskfile"
	"golang.org/x/sync/errgroup"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type TaskCmd string

const (
	NoCmd   TaskCmd = ""
	ListCmd TaskCmd = "list"
	// Summary is a command instead of an optional flag
	SummaryCmd TaskCmd = "summary"
	// Run runs a task
	RunCmd     TaskCmd = "run"
	VersionCmd TaskCmd = "version"

	// Number of output lines to respond with
	defaultNlines int = 5
)

var (
	errClosingWriter error = errors.New("cannot close closed writer")
	errClosedWriter  error = errors.New("cannot write to closed writer")

	// statusTaskSuccess is sent when task execution is finished.
	// Handy for long running tasks whose outputs are often transmitted late in time
	// Status codes from 3000 to 3999 are application specific
	statusTaskSuccess websocket.StatusCode = 3001
	statusTaskFailure websocket.StatusCode = 3002
)

type limitedWriter struct {
	closed bool

	c   *websocket.Conn
	ctx context.Context
	typ websocket.MessageType

	b *bytes.Buffer

	// Number of output lines to buffer and current lines written
	nLines, n int
}

// Returns a writer that flushes output to the underlying websocket on nLines writes
// Similar to a websocket writer: https://github.com/nhooyr/websocket/blob/3604edcb857415cb2c1213d63328cdcd738f2328/ws_js.go#L313
// This is essentially the websocket writer with modifications, so make sure you close it
func NewLimitedWriter(ctx context.Context, c *websocket.Conn, typ websocket.MessageType, nLines int) (limitedWriter, error) {
	if nLines == 0 {
		nLines = defaultNlines
	}
	return limitedWriter{
		c:   c,
		ctx: ctx,
		typ: typ,

		b: getBuf(),

		n:      0,
		nLines: nLines,
	}, nil
}

// Write implements the io.Writer interface
func (lw *limitedWriter) Write(p []byte) (int, error) {
	// fmt.Printf("woah calling Write: %s\n", p)
	if lw.closed {
		return 0, errClosedWriter
	}

	if lw.n >= lw.nLines {
		if err := lw.Flush(); err != nil {
			return 0, err
		}
	}
	lw.n += 1
	return lw.b.Write(p)
}

// Flush the output to the underlying websocket without closing it
func (lw *limitedWriter) Flush() error {
	log.Println("flushing output to websocket")
	// fmt.Printf("n: %d content: %s\n", lw.n, lw.b.Bytes())
	lw.n = 0
	err := lw.c.Write(lw.ctx, lw.typ, lw.b.Bytes())
	if err != nil {
		return fmt.Errorf("failed to flush output to websocket: %w", err)
	}
	lw.b.Reset()
	return nil
}

// FlushClose Closes the writer and flushes any output beforehand
// I avoided naming it Close() as this implementes the io.Closer interface
// and apparently Task calls Close() for unknown reasons after each command execution
func (lw *limitedWriter) FlushClose() error {
	if lw.closed {
		return errClosingWriter
	}
	lw.closed = true
	defer putBuf(lw.b)

	if err := lw.Flush(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}

// Options matching the command-line
type TaskOpts struct {
	Force   bool `json:"force"`
	Silent  bool `json:"silent"`
	Verbose int  `json:"verbose"`
	// Number of output lines to send back. Defaults to defaultNlines
	NLines int `json:"nLines"`
}

// The base request structure
type TaskReq struct {
	Command   TaskCmd  `json:"command"`
	Options   TaskOpts `json:"options"`
	TaskCalls []string `json:"tasks"`
}

// The base response structure
type TaskResp struct{}

type taskNameAndDesc struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type ListResp struct {
	Tasks []taskNameAndDesc
}

// Parses input request and runs the given command
func ParseAndRun(ctx context.Context, c *websocket.Conn, r TaskReq, s *BasicServer) error {

	// TODO: stdin is to be used later
	var stdin, stdout bytes.Buffer

	e := task.Executor{
		// Request options
		Force:   r.Options.Force,
		Verbose: r.Options.Verbose,
		Silent:  r.Options.Silent,
		Summary: r.Command == SummaryCmd,
		// Parallel: true,
		Color: false,

		// Task "server" options, often passed at initial invocation as server
		Entrypoint: s.Entrypoint,

		Stdin:  &stdin,
		Stdout: &stdout,
		Stderr: &stdout,
	}

	if err := e.Setup(); err != nil {
		log.Println(err)
		return err
	}

	switch r.Command {
	case ListCmd:
		t := listTasks(&e)
		if err := wsjson.Write(ctx, c, t); err != nil {
			log.Println(err)
			return err
		}

	case RunCmd:
		// limitedBufferedStdout is a buffered writer over the websocket writer
		limitedBufferedStdout, err := NewLimitedWriter(ctx, c, websocket.MessageText, r.Options.NLines)
		if err != nil {
			return fmt.Errorf("failed to initialize limitedWriter: %w", err)
		}
		e.Stdout = &limitedBufferedStdout
		e.Logger.Stdout = &limitedBufferedStdout
		e.Stderr = &limitedBufferedStdout
		e.Logger.Stderr = &limitedBufferedStdout

		// defer limitedBufferedStdout.Close()
		defer func() {
			if errClosure := limitedBufferedStdout.FlushClose(); errClosure != nil {
				log.Println("failed to close websocket custom writer: ", errClosure)
			}
		}()

		if err := runTasks(ctx, &e, r); err != nil {
			log.Println(err)
			return websocket.CloseError{Code: statusTaskFailure, Reason: err.Error()}
		}
		// TODO: Hack
		return websocket.CloseError{Code: statusTaskSuccess}

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

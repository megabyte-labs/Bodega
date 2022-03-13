package task

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AlecAivazis/survey/v2"
	tea "github.com/charmbracelet/bubbletea"
	"gitlab.com/megabyte-labs/go/cli/bodega/internal/compiler"
	compilerv2 "gitlab.com/megabyte-labs/go/cli/bodega/internal/compiler/v2"
	compilerv3 "gitlab.com/megabyte-labs/go/cli/bodega/internal/compiler/v3"
	"gitlab.com/megabyte-labs/go/cli/bodega/internal/execext"
	"gitlab.com/megabyte-labs/go/cli/bodega/internal/logger"
	"gitlab.com/megabyte-labs/go/cli/bodega/internal/output"
	"gitlab.com/megabyte-labs/go/cli/bodega/internal/summary"
	"gitlab.com/megabyte-labs/go/cli/bodega/internal/ui"
	"gitlab.com/megabyte-labs/go/cli/bodega/taskfile"
	"gitlab.com/megabyte-labs/go/cli/bodega/taskfile/read"
	"mvdan.cc/sh/v3/interp"

	"golang.org/x/sync/errgroup"
)

const (
	// MaximumTaskCall is the max number of times a task can be called.
	// This exists to prevent infinite loops on cyclic dependencies
	MaximumTaskCall = 100
)

// Executor executes a Taskfile
type Executor struct {
	Taskfile *taskfile.Taskfile

	Dir        string
	Entrypoint string
	Force      bool
	Watch      bool
	Verbose    int
	Silent     bool
	Dry        bool
	Summary    bool
	// Stop before every command execution. Might rename this later
	Debug       bool
	Parallel    bool
	Color       bool
	Concurrency int

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Logger      *logger.Logger
	FancyLogger *logger.FancyLogger
	Compiler    compiler.Compiler
	Output      output.Output
	OutputStyle string

	taskvars *taskfile.Vars

	concurrencySemaphore chan struct{}
	taskCallCount        map[string]*int32
	mkdirMutexMap        map[string]*sync.Mutex
	executionHashes      map[string]context.Context
	executionHashesMutex sync.Mutex
}

// Run runs Task
func (e *Executor) Run(ctx context.Context, calls ...taskfile.Call) error {
	// check if given tasks exist
	for _, c := range calls {
		if _, ok := e.Taskfile.Tasks[c.Task]; !ok {
			// FIXME: move to the main package
			e.ListTasksWithDesc()
			return &taskNotFoundError{taskName: c.Task}
		}
	}

	if e.Summary {
		var summaryBuilder strings.Builder
		// In order to keep compatability, each method in the summary packages prints
		// AND returns the output suited for markdown rendering. To prevent repeated
		// printing, output is temporarily redirected to /dev/null
		// if void := os.NewFile(0, os.DevNull); void != nil {
		defer func() { e.Logger.Stdout = e.Stdout }()
		// No need to close DevNull
		e.Logger.Stdout = execext.NewDevNull()

		if len(calls) > 1 {
			summaryBuilder.WriteString("# Tasks\n")
		}
		for i, c := range calls {
			compiledTask, err := e.FastCompiledTask(c)
			if err != nil {
				return err
			}
			summaryBuilder.WriteString(summary.PrintSpaceBetweenSummaries(e.Logger, i))
			summaryBuilder.WriteString(summary.PrintTask(e.Logger, compiledTask))
		}
		e.FancyLogger.Out(summaryBuilder.String())
		return nil
	}

	if e.Watch {
		return e.watchTasks(calls...)
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, c := range calls {
		c := c
		if e.Parallel {
			g.Go(func() error { return e.RunTask(ctx, c) })
		} else {
			if err := e.RunTask(ctx, c); err != nil {
				return err
			}
		}
	}
	return g.Wait()
}

// RunUI shows up a terminal interface. This must be called after Setup()
// Currently this is quite hacky
func (e *Executor) RunUI(ctx context.Context) error {
	// Temporary use-once channel receiving the task to run
	tChan := make(chan string, 1)
	model := ui.NewTasksModel(e.tasksWithDesc(), tChan)
	f := func(call taskfile.Call) error {
		if err := e.RunTask(ctx, call); err != nil {
			e.Logger.Errf(logger.Red, "%v", err)
			return err
		}
		return nil
	}

	if err := tea.NewProgram(model, tea.WithAltScreen()).Start(); err != nil {
		return err
	}
	deleteme := <-model.TChan
	if err := f(taskfile.Call{Task: deleteme}); err != nil {
		return err
	}
	return nil
}

// Setup setups Executor's internal state
func (e *Executor) Setup() error {
	var err error
	e.Taskfile, err = read.Taskfile(e.Dir, e.Entrypoint)
	if err != nil {
		return err
	}

	v, err := e.Taskfile.ParsedVersion()
	if err != nil {
		return err
	}

	if v < 3.0 {
		e.taskvars, err = read.Taskvars(e.Dir)
		if err != nil {
			return err
		}
	}

	if e.Stdin == nil {
		e.Stdin = os.Stdin
	}
	if e.Stdout == nil {
		e.Stdout = os.Stdout
	}
	if e.Stderr == nil {
		e.Stderr = os.Stderr
	}
	e.Logger = &logger.Logger{
		Stdout:  e.Stdout,
		Stderr:  e.Stderr,
		Verbose: e.Verbose,
		Color:   e.Color,
	}
	e.FancyLogger = logger.NewFancyLogger().SetStderr(e.Stderr).SetStdout(e.Stdout)

	if v < 2 {
		return fmt.Errorf(`task: Taskfile versions prior to v2 are not supported anymore`)
	}

	// consider as equal to the greater version if round
	if v == 2.0 {
		v = 2.6
	}
	if v == 3.0 {
		v = 3.7
	}

	if v > 3.7 {
		return fmt.Errorf(`task: Taskfile versions greater than v3.7 not implemented in the version of Task`)
	}

	// Color available only on v3
	if v < 3 {
		e.Logger.Color = false
	}

	if v < 3 {
		e.Compiler = &compilerv2.CompilerV2{
			Dir:          e.Dir,
			Taskvars:     e.taskvars,
			TaskfileVars: e.Taskfile.Vars,
			Expansions:   e.Taskfile.Expansions,
			Logger:       e.Logger,
		}
	} else {
		e.Compiler = &compilerv3.CompilerV3{
			Dir:          e.Dir,
			TaskfileEnv:  e.Taskfile.Env,
			TaskfileVars: e.Taskfile.Vars,
			Logger:       e.Logger,
		}
	}

	if v >= 3.0 {
		env, err := read.Dotenv(e.Compiler, e.Taskfile, e.Dir)
		if err != nil {
			return err
		}

		err = env.Range(func(key string, value taskfile.Var) error {
			if _, ok := e.Taskfile.Env.Mapping[key]; !ok {
				e.Taskfile.Env.Set(key, value)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	if v < 2.1 && e.Taskfile.Output != "" {
		return fmt.Errorf(`task: Taskfile option "output" is only available starting on Taskfile version v2.1`)
	}
	if v < 2.2 && e.Taskfile.Includes.Len() > 0 {
		return fmt.Errorf(`task: Including Taskfiles is only available starting on Taskfile version v2.2`)
	}
	if v >= 3.0 && e.Taskfile.Expansions > 2 {
		return fmt.Errorf(`task: The "expansions" setting is not available anymore on v3.0`)
	}

	if e.OutputStyle != "" {
		e.Taskfile.Output = e.OutputStyle
	}
	switch e.Taskfile.Output {
	case "", "interleaved":
		e.Output = output.Interleaved{}
	case "group":
		e.Output = output.Group{}
	case "prefixed":
		e.Output = output.Prefixed{}
	default:
		return fmt.Errorf(`task: output option "%s" not recognized`, e.Taskfile.Output)
	}

	if e.Taskfile.Method == "" {
		if v >= 3 {
			e.Taskfile.Method = "checksum"
		} else {
			e.Taskfile.Method = "timestamp"
		}
	}

	if v <= 2.1 {
		err := errors.New(`task: Taskfile option "ignore_error" is only available starting on Taskfile version v2.1`)

		for _, task := range e.Taskfile.Tasks {
			if task.IgnoreError {
				return err
			}
			for _, cmd := range task.Cmds {
				if cmd.IgnoreError {
					return err
				}
			}
		}
	}

	if v < 2.6 {
		for _, task := range e.Taskfile.Tasks {
			if len(task.Preconditions) > 0 {
				return errors.New(`task: Task option "preconditions" is only available starting on Taskfile version v2.6`)
			}
		}
	}

	if v < 3 {
		err := e.Taskfile.Includes.Range(func(_ string, taskfile taskfile.IncludedTaskfile) error {
			if taskfile.AdvancedImport {
				return errors.New(`task: Import with additional parameters is only available starting on Taskfile version v3`)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	if v < 3.7 {
		if e.Taskfile.Run != "" {
			return errors.New(`task: Setting the "run" type is only available starting on Taskfile version v3.7`)
		}

		for _, task := range e.Taskfile.Tasks {
			if task.Run != "" {
				return errors.New(`task: Setting the "run" type is only available starting on Taskfile version v3.7`)
			}
		}
	}

	if e.Taskfile.Run == "" {
		e.Taskfile.Run = "always"
	}

	e.executionHashes = make(map[string]context.Context)

	e.taskCallCount = make(map[string]*int32, len(e.Taskfile.Tasks))
	e.mkdirMutexMap = make(map[string]*sync.Mutex, len(e.Taskfile.Tasks))
	for k := range e.Taskfile.Tasks {
		e.taskCallCount[k] = new(int32)
		e.mkdirMutexMap[k] = &sync.Mutex{}
	}

	if e.Concurrency > 0 {
		e.concurrencySemaphore = make(chan struct{}, e.Concurrency)
	}
	return nil
}

// RunTask runs a task by its name
func (e *Executor) RunTask(ctx context.Context, call taskfile.Call) error {
	u := time.Now()
	t, err := e.CompiledTask(call)
	if err != nil {
		return err
	}

	if t.LogMsg != nil && t.LogMsg.Start != "" {
		e.Logger.Outf(logger.Magenta, t.LogMsg.Start)
	} else {
		e.Logger.VerboseErrf(logger.Magenta, `task: "%s" started`, call.Task)
	}

	if !e.Watch && atomic.AddInt32(e.taskCallCount[call.Task], 1) >= MaximumTaskCall {
		return &MaximumTaskCallExceededError{task: call.Task}
	}

	release := e.acquireConcurrencyLimit()
	defer release()

	return e.startExecution(ctx, t, func(ctx context.Context) error {
		if err := e.runDeps(ctx, t); err != nil {
			return err
		}

		// Do not execute task if preconditions are met or the task is up to date
		if !e.Force {
			if err := ctx.Err(); err != nil {
				return err
			}

			preCondMet, err := e.areTaskPreconditionsMet(ctx, t)
			if err != nil {
				return err
			}

			upToDate, err := e.isTaskUpToDate(ctx, t)
			if err != nil {
				return err
			}

			if upToDate && preCondMet {
				if !e.Silent {
					e.Logger.Errf(logger.Magenta, `task: Task "%s" is up to date`, t.Name())
				}
				return nil
			}
		}

		// By default, tasks will be executed in the directory where the Taskfile is located
		// unless the `dir` field is set
		if err := e.mkdir(t); err != nil {
			e.Logger.Errf(logger.Red, "task: cannot make directory %q: %v", t.Dir, err)
		}

		// NOTE: should prompts support shellRc ?
		if t.Prompt != nil {
			if err := e.runPrompt(ctx, t); err != nil {
				e.Logger.Errf(logger.Red, "task: prompt execution failed: %v", err)
			}
		}

		// Execute the initial shell script then pass the returned runner to each command
		var runner *interp.Runner = nil
		if t.ShellRc != "" {
			runner, err = execext.RunCommand(ctx, &execext.RunCommandOptions{
				Command: t.ShellRc,
				Dir:     t.Dir,
				Env:     getEnviron(t),
				Stdin:   e.Stdin,
				Stdout:  e.Stdout, // TODO: support Prefix
				Stderr:  e.Stderr,
			}, nil)
			if _, ok := execext.IsExitError(err); !ok {
				e.Logger.VerboseErrf(logger.Yellow, "task: [%s] error executing initial script: %v", t.Name(), err)
			}
		}

		// Execute all commands in the task
		for i := range t.Cmds {
			if t.Cmds[i].Defer {
				defer e.runDeferred(t, call, i)
				continue
			}

			if err := e.runCommand(ctx, t, call, i, runner); err != nil {
				if err2 := e.statusOnError(t); err2 != nil {
					e.Logger.VerboseErrf(logger.Yellow, "task: error cleaning status on error: %v", err2)
				}

				if _, ok := execext.IsExitError(err); ok && t.IgnoreError {
					e.Logger.VerboseErrf(logger.Yellow, "task: task error ignored: %v", err)
					continue
				}

				// Print exit status custom messages if found
				// Urgh this is a bit... ugly
				if t.LogMsg != nil && t.LogMsg.Error != nil {
					var customExitMsg bool
					code, ok := execext.IsExitError(err)
					if ok && t.LogMsg.Error.Codes != nil {
						for _, c := range t.LogMsg.Error.Codes {
							if c.Code == code {
								e.Logger.Outf(logger.Magenta, c.Message)
								customExitMsg = true
								break
							}
						}
					}
					// No custom exit codes defined and this is an exit code
					if !customExitMsg {
						e.Logger.Outf(logger.Magenta, t.LogMsg.Error.Default)
					}
				}
				return &taskRunError{t.Task, err}
			}
		}

		// The task execution time is accurate to a couple of milliseconds
		timeAfter := time.Now().Sub(u)
		e.Logger.VerboseErrf(logger.Magenta, `task: [%s] finished in %f seconds`, call.Task, timeAfter.Seconds())
		return nil
	})
}

func (e *Executor) mkdir(t *taskfile.Task) error {
	if t.Dir == "" {
		return nil
	}

	mutex := e.mkdirMutexMap[t.Task]
	mutex.Lock()
	defer mutex.Unlock()

	if _, err := os.Stat(t.Dir); os.IsNotExist(err) {
		if err := os.MkdirAll(t.Dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// runDeps runs all dependency tasks within task t
func (e *Executor) runDeps(ctx context.Context, t *taskfile.Task) error {
	g, ctx := errgroup.WithContext(ctx)

	reacquire := e.releaseConcurrencyLimit()
	defer reacquire()

	for _, d := range t.Deps {
		d := d

		g.Go(func() error {
			err := e.RunTask(ctx, taskfile.Call{Task: d.Task, Vars: d.Vars})
			if err != nil {
				return err
			}
			return nil
		})
	}

	return g.Wait()
}

func (e *Executor) runDeferred(t *taskfile.Task, call taskfile.Call, i int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := e.runCommand(ctx, t, call, i, nil); err != nil {
		e.Logger.VerboseErrf(logger.Yellow, `task: ignored error in deferred cmd: %s`, err.Error())
	}
}

func (e *Executor) runCommand(ctx context.Context, t *taskfile.Task, call taskfile.Call, i int, runner *interp.Runner) error {
	cmd := t.Cmds[i]

	switch {
	// A command can be interpreted as a taskfile.Task following this syntax:
	//  cmds:
	//    - task: task-to-be-called
	case cmd.Task != "":
		reacquire := e.releaseConcurrencyLimit()
		defer reacquire()

		err := e.RunTask(ctx, taskfile.Call{Task: cmd.Task, Vars: cmd.Vars})
		if err != nil {
			return err
		}
		return nil
	case cmd.Cmd != "":
		if e.Verbose != 0 || (!cmd.Silent && !t.Silent && !e.Taskfile.Silent && !e.Silent) {
			e.Logger.Errf(logger.Green, "task: [%s] %s", t.Name(), cmd.Cmd)
		}

		if e.Dry {
			return nil
		}

		outputWrapper := e.Output
		if t.Interactive {
			outputWrapper = output.Interleaved{}
		}
		stdOut := outputWrapper.WrapWriter(e.Stdout, t.Prefix)
		stdErr := outputWrapper.WrapWriter(e.Stderr, t.Prefix)

		defer func() {
			if _, ok := stdOut.(*os.File); !ok {
				if closer, ok := stdOut.(io.Closer); ok {
					closer.Close()
				}
			}
			if _, ok := stdErr.(*os.File); !ok {
				if closer, ok := stdErr.(io.Closer); ok {
					closer.Close()
				}
			}
		}()

		// Using mvdans/sh to run shell commands
		timeBefore := time.Now()
		_, err := execext.RunCommand(ctx, &execext.RunCommandOptions{
			Command: cmd.Cmd,
			Dir:     t.Dir,
			Env:     getEnviron(t),
			Debug:   e.Debug,
			Stdin:   e.Stdin,
			Stdout:  stdOut,
			Stderr:  stdErr,
		}, runner)
		timeAfter := time.Now().Sub(timeBefore)
		e.Logger.DebugOutf(logger.Cyan, "task: [%s] command %s took %v ms", t.Name(), cmd.Cmd, timeAfter.Milliseconds())
		if _, ok := execext.IsExitError(err); ok && cmd.IgnoreError {
			e.Logger.VerboseErrf(logger.Yellow, "task: [%s] command error ignored: %v", t.Name(), err)
			return nil
		}
		return err
	default:
		return nil
	}
}

func getEnviron(t *taskfile.Task) []string {
	if t.Env == nil {
		return nil
	}

	environ := os.Environ()

	for k, v := range t.Env.ToCacheMap() {
		str, isString := v.(string)
		if !isString {
			continue
		}

		if _, alreadySet := os.LookupEnv(k); alreadySet {
			continue
		}

		environ = append(environ, fmt.Sprintf("%s=%s", k, str))
	}

	return environ
}

// startExecution is a helper fucntion used inside Executor.RunTask to execute commands.
// It decides whether to run task t or not based on t.Run value
func (e *Executor) startExecution(ctx context.Context, t *taskfile.Task, execute func(ctx context.Context) error) error {
	h, err := e.GetHash(t)
	if err != nil {
		return err
	}

	// Always run the task
	if h == "" {
		return execute(ctx)
	}

	// Persist running the task across reboots if "run" property is "once"
	// Note that if a task ran before with run_once_system set to true, then
	// it should not be removed in order for Task to recognize that.
	if t.Run == "once" && t.RunOnceSystem {
		c, _ := os.UserCacheDir()
		f := filepath.Join(c, "bodega", h)
		// TODO: os.Stat might return false positinves or suffer from TOCTOU race condition
		if _, err := os.Stat(f); os.IsNotExist(err) {
			if err = os.MkdirAll(filepath.Join(c, "bodega"), 0o755); err == nil {
				e.Logger.Errf(logger.Red, "%v", err)
			}
			if _, err := os.Create(f); err != nil {
				e.Logger.Errf(logger.Red, "task: error writing file: %v", err)
			}
		} else {
			e.executionHashesMutex.Lock()
			dummyCtx, cancel := context.WithCancel(context.TODO())
			e.executionHashes[h] = dummyCtx
			e.executionHashesMutex.Unlock()
			cancel()
		}
	}

	e.executionHashesMutex.Lock()
	otherExecutionCtx, ok := e.executionHashes[h]

	if ok {
		e.executionHashesMutex.Unlock()
		e.Logger.VerboseErrf(logger.Magenta, "task: skipping execution of task: %s", h)
		<-otherExecutionCtx.Done()
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	e.executionHashes[h] = ctx
	e.executionHashesMutex.Unlock()

	return execute(ctx)
}

// Run the Prompt setup based on the YAML configuration
func (e *Executor) runPrompt(ctx context.Context, t *taskfile.Task) error {
	// Set the Answer Task
	if t.Prompt.Answer != nil {
		t.Prompt.Answer.Task = fmt.Sprintf("%s.%s", t.Task, "answer")
		e.Taskfile.Tasks[t.Prompt.Answer.Task] = t.Prompt.Answer
		e.taskCallCount[t.Prompt.Answer.Task] = new(int32)
		e.mkdirMutexMap[t.Prompt.Answer.Task] = &sync.Mutex{}
	}

	var (
		prompt survey.Prompt
		// Used by a couple of prompt types
		inputText string
	)

	funcValidateAndRunAnswer := func(answer string) error {
		if t.Vars != nil {
			t.Vars.Set("ANSWER", taskfile.Var{Static: answer})
		} else {
			v := taskfile.Vars{}
			v.Set("ANSWER", taskfile.Var{Static: answer})
			t.Vars = &v
		}
		// FIXME: do we have to compile the whole task ?
		temp, _ := e.CompiledTask(taskfile.Call{Task: t.Task, Vars: t.Vars})
		if temp.Prompt.Validate != nil {
			_, err := execext.RunCommand(ctx, &execext.RunCommandOptions{
				Command: temp.Prompt.Validate.Sh,
				Env:     getEnviron(t),
			}, nil)
			if err != nil {
				return fmt.Errorf("validation failed: %v", err)
			}
		}
		return e.RunTask(ctx, taskfile.Call{Task: t.Prompt.Answer.Task, Vars: t.Vars})
	}

	// A common procedure used by Select and MultiSelect prompt types.
	selectItemsFunc := func(isMultiSelection bool) error {
		var (
			options  []string
			selected string
			// Used to represent a set to store unique options
			optionsMap = make(map[string]struct{}, len(t.Prompt.Options.Values))
		)

		if t.Prompt.Options.JsonArr != "" {
			var choices []string
			if err := json.Unmarshal([]byte(t.Prompt.Options.JsonArr), &choices); err == nil {
				for _, c := range choices {
					t.Prompt.Options.Values = append(t.Prompt.Options.Values, taskfile.ValueType{Value: c})
				}
			}

		}

		// Parse `prompt` options
		for _, option := range t.Prompt.Options.Values {
			if option.Value == "" && option.Msg != nil {
				if option.Msg.Value != "" {
					option.Value = option.Msg.Value
				} else {
					// Execute the command(s) within "sh:" field and capture the output
					var out bytes.Buffer
					_, err := execext.RunCommand(context.Background(), &execext.RunCommandOptions{
						Command: option.Msg.Sh,
						Dir:     t.Dir,
						Env:     getEnviron(t),
						Stdout:  &out,
					}, nil)
					if err != nil {
						e.Logger.VerboseOutf(logger.Yellow, "command %s at prompt %s exited abnormally: %v", option.Msg.Sh, t.Name(), err)
						return err
					}
					option.Value = strings.TrimRight(out.String(), " \n")
				}
			}
			if _, ok := optionsMap[option.Value]; !ok {
				optionsMap[option.Value] = struct{}{}
				options = append(options, option.Value)
			}
		}

		// Use either Select or MultiSelect and validate each selection
		if isMultiSelection {
			var m []string
			prompt = &survey.MultiSelect{
				Message: t.Prompt.Message,
				Options: options,
			}
			survey.AskOne(prompt, &m)
			for i := range m {
				// Build a string to be looped on for {{.ANSWER}}
				selected += "\"" + m[i] + "\" "
			}
		} else {
			prompt = &survey.Select{
				Message: t.Prompt.Message,
				Options: options,
			}
			survey.AskOne(prompt, &selected)
		}
		// The answer task runs only once, even for multi_select
		if err := funcValidateAndRunAnswer(selected); err != nil {
			return err
		}

		return nil
	}
	switch t.Prompt.Type {
	case "input":
		prompt := &survey.Input{
			Message: t.Prompt.Message,
		}
		survey.AskOne(prompt, &inputText)
		if err := funcValidateAndRunAnswer(inputText); err != nil {
			return err
		}

	case "multiline":
		prompt = &survey.Multiline{
			Message: t.Prompt.Message,
		}
		survey.AskOne(prompt, &inputText)

		if err := funcValidateAndRunAnswer(inputText); err != nil {
			return err
		}
	case "password":
		prompt = &survey.Password{
			Message: t.Prompt.Message,
		}
		survey.AskOne(prompt, &inputText)

		if err := funcValidateAndRunAnswer(inputText); err != nil {
			return err
		}

	case "confirm":
		var yes bool
		prompt = &survey.Confirm{
			Message: t.Prompt.Message,
		}
		survey.AskOne(prompt, &yes)
		if yes {
			if err := e.RunTask(ctx, taskfile.Call{Task: t.Prompt.Answer.Task}); err != nil {
				return err
			}
		}

	case "select":
		if err := selectItemsFunc(false); err != nil {
			return err
		}

	case "multi_select":
		if err := selectItemsFunc(true); err != nil {
			return err
		}

	default:
		e.Logger.Errf(logger.Cyan, "`Invalid option`")
	}
	return nil
}

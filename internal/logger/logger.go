package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
)

type (
	Color     func() PrintFunc
	PrintFunc func(io.Writer, string, ...interface{})
)

func Default() PrintFunc { return color.New(color.Reset).FprintfFunc() }
func Blue() PrintFunc    { return color.New(color.FgBlue).FprintfFunc() }
func Green() PrintFunc   { return color.New(color.FgGreen).FprintfFunc() }
func Cyan() PrintFunc    { return color.New(color.FgCyan).FprintfFunc() }
func Yellow() PrintFunc  { return color.New(color.FgYellow).FprintfFunc() }
func Magenta() PrintFunc { return color.New(color.FgMagenta).FprintfFunc() }
func Red() PrintFunc     { return color.New(color.FgRed).FprintfFunc() }

const (
	verbosityLevelNone int = iota
	verbosityLevelInfo
	verbosityLevelDebug
)

// Logger is just a wrapper that prints stuff to STDOUT or STDERR,
// with optional color.
type Logger struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Verbose int
	Color   bool
}

// A logger which renders markdown with fancy colors thanks to charmbracelet/glamour
// TODO: refactor to one logger interface
type FancyLogger struct {
	stdout io.Writer
	stderr io.Writer
	// TODO: Verbose should indicate with verbosity levels
	Verbose int
	// TODO: keeping this for now although it's no use
	Color   bool
	glamour *glamour.TermRenderer
}

// Initializes and returns a fancy logger backed with charmbracelet/glamour
func NewFancyLogger() *FancyLogger {
	fl, err := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing fancy logger: %v", err)
	}
	return &FancyLogger{
		stdout:  os.Stdout,
		stderr:  os.Stderr,
		Verbose: verbosityLevelInfo,
		Color:   false,
		glamour: fl,
	}
}

func (fl *FancyLogger) SetStderr(fd io.Writer) *FancyLogger {
	fl.stderr = fd
	return fl
}

func (fl *FancyLogger) SetStdout(fd io.Writer) *FancyLogger {
	fl.stdout = fd
	return fl
}

func (fl *FancyLogger) Err(s string) {
	out, err := fl.glamour.Render(s)
	if err != nil {
		fmt.Fprint(fl.stderr, err)
		return
	}
	fmt.Fprint(fl.stderr, out)
}

func (fl *FancyLogger) Out(s string) {
	out, err := fl.glamour.Render(s)
	if err != nil {
		fmt.Fprint(fl.stderr, err)
		return
	}
	fmt.Fprint(fl.stdout, out)
}

// Outf prints stuff to STDOUT.
func (l *Logger) Outf(color Color, s string, args ...interface{}) {
	if len(args) == 0 {
		s, args = "%s", []interface{}{s}
	}
	if !l.Color {
		color = Default
	}
	print := color()
	print(l.Stdout, s+"\n", args...)
}

// VerboseOutf prints stuff to STDOUT if verbose mode is enabled.
func (l *Logger) VerboseOutf(color Color, s string, args ...interface{}) {
	if l.Verbose >= verbosityLevelInfo {
		l.Outf(color, s, args...)
	}
}

func (l *Logger) DebugOutf(color Color, s string, args ...interface{}) {
	if l.Verbose >= verbosityLevelDebug {
		l.Outf(color, s, args...)
	}
}

// Errf prints stuff to STDERR.
func (l *Logger) Errf(color Color, s string, args ...interface{}) {
	if len(args) == 0 {
		s, args = "%s", []interface{}{s}
	}
	if !l.Color {
		color = Default
	}
	print := color()
	print(l.Stderr, s+"\n", args...)
}

// VerboseErrf prints stuff to STDERR if verbose mode is enabled.
func (l *Logger) VerboseErrf(color Color, s string, args ...interface{}) {
	if l.Verbose >= verbosityLevelInfo {
		l.Errf(color, s, args...)
	}
}

func (l *Logger) DebugErrf(color Color, s string, args ...interface{}) {
	if l.Verbose >= verbosityLevelDebug {
		l.Errf(color, s, args...)
	}
}

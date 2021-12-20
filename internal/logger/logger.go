package logger

import (
	"io"

	"github.com/fatih/color"
)

type Color func() PrintFunc
type PrintFunc func(io.Writer, string, ...interface{})

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

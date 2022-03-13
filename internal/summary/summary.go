package summary

import (
	"strings"

	"gitlab.com/megabyte-labs/go/cli/bodega/internal/logger"
	"gitlab.com/megabyte-labs/go/cli/bodega/taskfile"
)

func PrintTasks(l *logger.Logger, t *taskfile.Taskfile, c []taskfile.Call) {
	for i, call := range c {
		PrintSpaceBetweenSummaries(l, i)
		PrintTask(l, t.Tasks[call.Task])
	}
}

func PrintSpaceBetweenSummaries(l *logger.Logger, i int) string {
	spaceRequired := i > 0
	if !spaceRequired {
		return "\n"
	}

	l.Outf(logger.Default, "")
	l.Outf(logger.Default, "")
	return "\n---\n"
}

// Prints a summarized form of a task t. In addition, returns
// the printed text suitable for markdown rendering
func PrintTask(l *logger.Logger, t *taskfile.Task) string {
	var out string
	out += printTaskName(l, t)
	out += printTaskDescribingText(t, l)
	out += printTaskDependencies(l, t)
	out += printTaskCommands(l, t)
	return out
}

func printTaskDescribingText(t *taskfile.Task, l *logger.Logger) string {
	if hasSummary(t) {
		return printTaskSummary(l, t)
	} else if hasDescription(t) {
		return printTaskDescription(l, t)
	} else {
		return printNoDescriptionOrSummary(l)
	}
}

func hasSummary(t *taskfile.Task) bool {
	return t.Summary != ""
}

func printTaskSummary(l *logger.Logger, t *taskfile.Task) string {
	lines := strings.Split(t.Summary, "\n")
	out := ""
	for i, line := range lines {
		notLastLine := i+1 < len(lines)
		if notLastLine || line != "" {
			l.Outf(logger.Default, line)
			out += "\n" + line + "\n"
		}
	}
	return out
}

func printTaskName(l *logger.Logger, t *taskfile.Task) string {
	out := "## Task `" + t.Name() + "`\n"
	l.Outf(logger.Default, "task: %s", t.Name())
	l.Outf(logger.Default, "")
	return out
}

func hasDescription(t *taskfile.Task) bool {
	return t.Desc != ""
}

func printTaskDescription(l *logger.Logger, t *taskfile.Task) string {
	l.Outf(logger.Default, t.Desc)
	return t.Desc + "\n"
}

func printNoDescriptionOrSummary(l *logger.Logger) string {
	out := "(task does not have description or summary)"
	l.Outf(logger.Default, out)
	return out + "\n"
}

func printTaskDependencies(l *logger.Logger, t *taskfile.Task) string {
	if len(t.Deps) == 0 {
		return ""
	}

	out := "\n## Dependencies\n"
	l.Outf(logger.Default, "")
	l.Outf(logger.Default, "dependencies:")

	for _, d := range t.Deps {
		l.Outf(logger.Default, " - %s", d.Task)
		out += "- " + d.Task + "\n"
	}
	return out
}

func printTaskCommands(l *logger.Logger, t *taskfile.Task) string {
	if len(t.Cmds) == 0 {
		return ""
	}

	out := "\n## Commands\n"
	l.Outf(logger.Default, "")
	l.Outf(logger.Default, "commands:")
	for _, c := range t.Cmds {
		isCommand := c.Cmd != ""
		if isCommand {
			l.Outf(logger.Default, " - %s", c.Cmd)
			if strings.Contains(c.Cmd, "\n") {
				// Assume a code block with indendation embedded
				// This is quite hacky if you ask me
				out += "- \n```bash\n" + c.Cmd + "```\n"
			} else {
				out += "- " + c.Cmd + "\n"
			}
		} else {
			l.Outf(logger.Default, " - Task: %s", c.Task)
			out += "- `" + c.Task + "`\n"
		}
	}
	return out
}

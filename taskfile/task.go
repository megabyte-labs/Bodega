package taskfile

// Tasks represents a group of tasks
type Tasks map[string]*Task

// Task represents a task
type Task struct {
	Task          string
	Alias         string
	ShellRc       string
	Cmds          []*Cmd
	Deps          []*Dep
	Label         string
	LogMsg        *LogMsg
	Desc          string
	Summary       string
	Sources       []string
	Generates     []string
	Status        []string
	Preconditions []*Precondition
	Dir           string
	Vars          *Vars
	Env           *Vars
	Silent        bool
	Interactive   bool
	Method        string
	Prefix        string
	IgnoreError   bool
	Run           string
	// TODO: Hide should be bool but we want Go templates
	Hide   string
	Prompt *Prompt
}

func (t *Task) Name() string {
	if t.Label != "" {
		return t.Label
	}
	return t.Task
}

func (t *Task) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cmd Cmd
	if err := unmarshal(&cmd); err == nil && cmd.Cmd != "" {
		t.Cmds = append(t.Cmds, &cmd)
		return nil
	}

	var cmds []*Cmd
	if err := unmarshal(&cmds); err == nil && len(cmds) > 0 {
		t.Cmds = cmds
		return nil
	}

	var task struct {
		Alias         string
		ShellRc       string `yaml:"shell_rc"`
		Cmds          []*Cmd
		Deps          []*Dep
		Label         string
		LogMsg        *LogMsg `yaml:"log"`
		Desc          string
		Summary       string
		Sources       []string
		Generates     []string
		Status        []string
		Preconditions []*Precondition
		Dir           string
		Vars          *Vars
		Env           *Vars
		Silent        bool
		Interactive   bool
		Method        string
		Prefix        string
		IgnoreError   bool `yaml:"ignore_error"`
		Run           string
		Hide          string
		Prompt        *Prompt
	}
	if err := unmarshal(&task); err != nil {
		return err
	}
	t.ShellRc = task.ShellRc
	t.Cmds = task.Cmds
	t.Deps = task.Deps
	t.Alias = task.Alias
	t.Label = task.Label
	t.LogMsg = task.LogMsg
	t.Desc = task.Desc
	t.Summary = task.Summary
	t.Sources = task.Sources
	t.Generates = task.Generates
	t.Status = task.Status
	t.Preconditions = task.Preconditions
	t.Dir = task.Dir
	t.Vars = task.Vars
	t.Env = task.Env
	t.Silent = task.Silent
	t.Interactive = task.Interactive
	t.Method = task.Method
	t.Prefix = task.Prefix
	t.IgnoreError = task.IgnoreError
	t.Run = task.Run
	t.Hide = task.Hide
	t.Prompt = task.Prompt
	return nil
}

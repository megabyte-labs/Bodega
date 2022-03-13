package taskfile

// Call is the parameter to a task call
type Call struct {
	Task string
	Vars *Vars
}

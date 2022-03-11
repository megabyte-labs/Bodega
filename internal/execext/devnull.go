package execext

import (
	"io"
)

var _ io.ReadWriteCloser = devNull{}

type devNull struct{}

// Returns a pointer to a ReadWriteCloser structure simulating /dev/null
func NewDevNull() *devNull {
	return &devNull{}
}

func (devNull) Read(p []byte) (int, error)  { return 0, io.EOF }
func (devNull) Write(p []byte) (int, error) { return len(p), nil }
func (devNull) Close() error                { return nil }

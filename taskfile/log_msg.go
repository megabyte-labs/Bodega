package taskfile

// Message to output on task error. Ideally, the end message should not be used if this is set
type LogMsgError struct {
		// Default message to output
		Default string
		// Custom messages for each status code
		Codes []struct {
			Code    uint8
			Message string
		}
}

// LogMsg represent a custom log message to print upon start/end of a task
type LogMsg struct {
	// Message to output on task start
	Start string
	Error *LogMsgError
	// Message to output on task end
	Success string
}

func (l *LogMsg) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var msgs struct {
		Start   string
		Error   *LogMsgError
		Success string
	}
	if err := unmarshal(&msgs); err != nil {
		return err
	}
	l.Start = msgs.Start
	l.Success = msgs.Success
	l.Error = msgs.Error
	return nil
}

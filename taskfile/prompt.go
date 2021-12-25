package taskfile

type ValueType struct {
	Value string
	Msg   *MsgType
}

type MsgType struct {
	Value string
	Sh    string
}

type Prompt struct {
	Type     string
	Message  string
	Options  []*ValueType
	Validate *MsgType
	Answer   *Task
}

// UnmarshalYAML implements yaml.Unmarshaler interface.
func (m *MsgType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cmd string

	// If a string is passed as message value, store in Value
	if err := unmarshal(&cmd); err == nil {
		m.Value = cmd
		return nil
	}

	// If a shell command is passed as message value, parse shell command and store value in Value
	var msgtype struct {
		Sh string
	}

	if err := unmarshal(&msgtype); err != nil {
		return err
	}

	m.Sh = msgtype.Sh
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface.
func (v *ValueType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cmd string

	if err := unmarshal(&cmd); err == nil {
		v.Value = cmd
		v.Msg = nil
		return nil
	}

	// An option value can contain a string, or a computed shell property
	var valuetype struct {
		Msg *MsgType
	}

	if err := unmarshal(&valuetype); err != nil {
		return err
	}

	v.Value = ""
	v.Msg = valuetype.Msg

	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface.
func (p *Prompt) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cmd string

	if err := unmarshal(&cmd); err == nil {
		return nil
	}

	var prompt struct {
		Type     string
		Message  string
		Options  []*ValueType
		Validate *MsgType
		Answer   *Task
	}

	if err := unmarshal(&prompt); err != nil {
		return err
	}

	p.Type = prompt.Type
	p.Message = prompt.Message
	p.Options = prompt.Options
	p.Validate = prompt.Validate
	p.Answer = prompt.Answer

	return nil
}

package taskfile

type ValueType struct {
	Value string
	Msg   *MsgType
}

type MsgType struct {
	Value string
	Sh    string
}

type Options struct {
	// Parse options from this json array string
	JsonArr string
	Values  []ValueType
}

type Prompt struct {
	Type     string
	Message  string
	Options  Options
	Validate *MsgType
	Answer   *Task
}

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

// UnmarshalYAML requires a pointer receiver even if the structure passed to
// marshal is not a pointer
func (o *Options) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var jArr string
	if err := unmarshal(&jArr); err == nil {
		o.JsonArr = jArr
		return err
	}

	var opts []ValueType
	if err := unmarshal(&opts); err != nil {
		return err
	}

	o.Values = opts
	return nil
}

func (p *Prompt) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var prompt struct {
		Type     string
		Message  string
		Options  Options
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

package toml

const (
	OptionNewLineLF     = "\n"
	OptionNewLineCR     = "\r"
	OptionNewLineCRLF   = "\r\n"
	OptionIndentDefault = "  "
)

// Option callback for connection option
type Option func(*Opts) error

func SetNewLineType(val string) Option {
	return func(t *Opts) error {
		t.lineEnding = val
		return nil
	}
}

func SetIndent(val string) Option {
	return func(t *Opts) error {
		t.indent = val
		return nil
	}
}

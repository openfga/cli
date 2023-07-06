package output

import "errors"

// Format describes the output formatting.
type Format string

const (
	FlagMessage = `output format.  Allowable value: "standard-json", "pretty-json"`
	FlagName    = "format"
)

var ErrUnknownFormat = errors.New(`must be one of \"standard-json\", \"pretty-json\"`)

const (
	StandardJSON Format = "standard-json"
	PrettyJSON   Format = "pretty-json"
)

// String is used both by fmt.Print and by Cobra in help text.
func (f *Format) String() string {
	return string(*f)
}

// Set validates the output.
func (f *Format) Set(v string) error {
	switch v {
	case "standard-json", "pretty-json":
		*f = Format(v)

		return nil
	default:
		return ErrUnknownFormat
	}
}

// Type provides help text.
func (f *Format) Type() string {
	return "format"
}

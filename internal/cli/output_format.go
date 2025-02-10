package cli

import (
	"encoding"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/aereal/iter/seq2"
	cli "github.com/urfave/cli/v3"
)

type InvalidOutputFormatError struct {
	V string
}

func (e *InvalidOutputFormatError) Error() string {
	return fmt.Sprintf("invalid output format: %q", e.V)
}

func (e *InvalidOutputFormatError) Is(other error) bool {
	otherErr := new(InvalidOutputFormatError)
	if !errors.As(other, &otherErr) {
		return false
	}
	return otherErr.V == e.V
}

type OutputFormat int

var (
	_ encoding.TextMarshaler   = (*OutputFormat)(nil)
	_ encoding.TextUnmarshaler = (*OutputFormat)(nil)
)

const (
	OutputFormatJSON OutputFormat = iota
	OutputFormatJSONPretty
)

var (
	of2str = map[OutputFormat]string{
		OutputFormatJSON:       "json",
		OutputFormatJSONPretty: "json.pretty",
	}
	str2of               = maps.Collect(seq2.Flip(maps.All(of2str)))
	definedOutputFormats = slices.Sorted(maps.Keys(of2str))
)

func AvailableOutputFormatValues() []OutputFormat {
	return definedOutputFormats
}

func (of OutputFormat) defined() bool {
	_, ok := of2str[of]
	return ok
}

func (of OutputFormat) MarshalText() ([]byte, error) {
	if !of.defined() {
		return nil, &InvalidOutputFormatError{V: of.String()}
	}
	return []byte(of.String()), nil
}

func (of *OutputFormat) UnmarshalText(b []byte) error {
	input := string(b)
	var ok bool
	*of, ok = str2of[input]
	if !ok {
		return &InvalidOutputFormatError{V: input}
	}
	return nil
}

func (of OutputFormat) String() string {
	s, ok := of2str[of]
	if !ok {
		return fmt.Sprintf("OutputFormat(%d)", of)
	}
	return s
}

type outputFormatValue OutputFormat

var _ cli.Value = (*outputFormatValue)(nil)

func (v outputFormatValue) String() string {
	return (OutputFormat)(v).String()
}

func (of *outputFormatValue) Set(v string) error {
	return (*OutputFormat)(of).UnmarshalText([]byte(v))
}

func (of *outputFormatValue) Get() any { return (OutputFormat)(*of) }

type outputFormatCreator struct{}

var _ cli.ValueCreator[OutputFormat, cli.NoConfig] = (*outputFormatCreator)(nil)

func (outputFormatCreator) Create(v OutputFormat, ref *OutputFormat, _ cli.NoConfig) cli.Value { //nolint:ireturn
	*ref = v
	return (*outputFormatValue)(ref)
}

func (outputFormatCreator) ToString(v OutputFormat) string {
	return v.String()
}

package cli_test

import (
	"errors"
	"testing"

	"github.com/aereal/frontier/internal/cli"
	"github.com/google/go-cmp/cmp"
)

func TestInvalidOutputFormatError_Is(t *testing.T) {
	testCases := []struct {
		name string
		lhs  error
		rhs  error
		want bool
	}{
		{
			name: "same type, same format",
			lhs:  &cli.InvalidOutputFormatError{V: "unknown"},
			rhs:  &cli.InvalidOutputFormatError{V: "unknown"},
			want: true,
		},
		{
			name: "same type, different format",
			lhs:  &cli.InvalidOutputFormatError{V: "unknown"},
			rhs:  &cli.InvalidOutputFormatError{V: "invalid"},
			want: false,
		},
		{
			name: "different type",
			lhs:  &cli.InvalidOutputFormatError{V: "unknown"},
			rhs:  &literalError{"oops"},
			want: false,
		},
		{
			name: "other is nil",
			lhs:  &cli.InvalidOutputFormatError{V: "unknown"},
			rhs:  nil,
			want: false,
		},
		{
			name: "self is nil",
			lhs:  nil,
			rhs:  &cli.InvalidOutputFormatError{V: "unknown"},
			want: false,
		},
		{
			name: "nil",
			lhs:  nil,
			rhs:  nil,
			want: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := errors.Is(tc.lhs, tc.rhs)
			if got != tc.want {
				t.Logf("lhs: %T %s", tc.lhs, tc.lhs)
				t.Logf("rhs: %T %s", tc.rhs, tc.rhs)
				t.Errorf("want=%v got=%v", tc.want, got)
			}
		})
	}
}

func TestOutputFormat_marshal(t *testing.T) {
	testCases := []struct {
		name    string
		format  cli.OutputFormat
		wantVal string
		wantErr error
	}{
		{
			name:    "ok/json",
			format:  cli.OutputFormatJSON,
			wantVal: "json",
			wantErr: nil,
		},
		{
			name:    "empty",
			format:  cli.OutputFormat(-1),
			wantErr: &cli.InvalidOutputFormatError{V: "OutputFormat(-1)"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotVal, gotErr := tc.format.MarshalText()
			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("error:\n\twant: %T %s\n\t got: %T %s", tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if gotErr != nil {
				return
			}
			if got := string(gotVal); got != tc.wantVal {
				t.Errorf("marshaled value:\n\twant: %q\n\t got: %q", tc.wantVal, got)
			}
		})
	}
}

func TestOutputFormat_unmarshal(t *testing.T) {
	testCases := []struct {
		name    string
		input   []byte
		wantVal cli.OutputFormat
		wantErr error
	}{
		{
			name:    "json",
			input:   []byte("json"),
			wantVal: cli.OutputFormatJSON,
		},
		{
			name:    "json pretty",
			input:   []byte("json.pretty"),
			wantVal: cli.OutputFormatJSONPretty,
		},
		{
			name:    "unknown value",
			input:   []byte("unknown"),
			wantErr: &cli.InvalidOutputFormatError{V: "unknown"},
		},
		{
			name:    "empty",
			input:   nil,
			wantErr: &cli.InvalidOutputFormatError{V: ""},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var gotVal cli.OutputFormat
			gotErr := (&gotVal).UnmarshalText(tc.input)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("error:\n\twant: %T %s\n\t got: %T %s", tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if gotErr != nil {
				return
			}
			if diff := cmp.Diff(tc.wantVal, gotVal); diff != "" {
				t.Errorf("value (-want, +got):\n%s", diff)
			}
		})
	}
}

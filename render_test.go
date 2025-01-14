package frontier_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/aereal/frontier"
	"github.com/google/go-cmp/cmp"
)

var wantConfig = `name: test-func
code:
  path: ./testdata/fn.js
config:
  comment: blah blah
  runtime: cloudfront-js-1.0
`

func TestRenderer_Render(t *testing.T) {
	testCases := []struct {
		name       string
		wantOutput string
		wantErr    error
	}{
		{name: "ok", wantOutput: wantConfig, wantErr: nil},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			if deadline, ok := t.Deadline(); ok {
				ctx, cancel = context.WithDeadline(ctx, deadline)
			}
			defer cancel()

			buf := new(bytes.Buffer)
			renderer := frontier.NewRenderer("./testdata/config.yml", buf)
			gotErr := renderer.Render(ctx)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("want error: %s\n got error: %s", tc.wantErr, gotErr)
			}
			if diff := cmp.Diff(tc.wantOutput, buf.String()); diff != "" {
				t.Errorf("output (-want, +got):\n%s", diff)
			}
		})
	}
}

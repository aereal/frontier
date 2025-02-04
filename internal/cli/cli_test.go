package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aereal/frontier"
	"github.com/aereal/frontier/internal/cli"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	gomock "go.uber.org/mock/gomock"
)

func TestApp_Run(t *testing.T) {
	testdataDir := "../../testdata"
	configPath := filepath.Join(testdataDir, "config.yml")
	tcs := []testSubcommandArgs{
		{
			args: []string{"deploy", "--config", configPath},
			expectDeploy: func(m *mockWithLogger[*cli.MockDeployController]) {
				m.M.EXPECT().
					Deploy(gomock.Any(), configPath, true).
					Return(nil).
					Times(1)
			},
		},
		{
			args: []string{"deploy", "--config", configPath, "--no-publish"},
			expectDeploy: func(m *mockWithLogger[*cli.MockDeployController]) {
				m.M.EXPECT().
					Deploy(gomock.Any(), configPath, false).
					Return(nil).
					Times(1)
			},
		},
		{
			args: []string{"deploy", "--config", configPath, "--publish"},
			expectDeploy: func(m *mockWithLogger[*cli.MockDeployController]) {
				m.M.EXPECT().
					Deploy(gomock.Any(), configPath, true).
					Return(nil).
					Times(1)
			},
		},
		{
			args:   []string{"deploy", "--config", configPath, "--publish", "--no-publish"},
			expect: testSubommandExpectation{err: &literalError{"option publish cannot be set along with option no-publish"}},
		},
		{
			args:   []string{"import", "--config", configPath, "--name", ""},
			expect: testSubommandExpectation{err: cli.ErrFunctionNameRequired},
		},
		{
			args:   []string{"import", "--config", configPath, "--name", "test-fn", "--function-path", ""},
			expect: testSubommandExpectation{err: cli.ErrFunctionPathRequired},
		},
		{
			args:   []string{"import", "--name", "test-fn", "--config", ""},
			expect: testSubommandExpectation{err: cli.ErrConfigPathRequired},
		},
		{
			args: []string{"render", "--config", configPath},
			expectRender: func(m *mockWithLogger[*cli.MockRenderController]) {
				m.M.EXPECT().
					Render(gomock.Any(), configPath, gomock.Any()).
					Return(nil).
					Times(1)
			},
		},
		{
			args: []string{"--log-level", "DEBUG", "render", "--config", configPath},
			expectRender: func(m *mockWithLogger[*cli.MockRenderController]) {
				m.M.EXPECT().
					Render(gomock.Any(), configPath, gomock.Any()).
					Return(nil).
					Times(1)
			},
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			testSubcommand(t, tc)
		})
	}
}

func TestApp_Run_import(t *testing.T) {
	wantFunctionBody := "console.log(1);\n"
	wantConfig := "name: test-fn\ncode:\n  path: imported.js\n"
	tmpDir := "../../testdata/tmp"
	importedConfigPath := filepath.Join(tmpDir, "imported.yml")
	importedFunctionPath := filepath.Join(tmpDir, "imported.js")
	args := testSubcommandArgs{
		args: []string{"import", "--config", importedConfigPath, "--name", "test-fn", "--function-path", importedFunctionPath},
		expectImport: func(m *mockWithLogger[*cli.MockImportController]) {
			m.M.EXPECT().
				Import(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, functionName string, configOutput io.Writer, functionFile *frontier.WritableFile) error {
					fmt.Fprint(functionFile.Writer, wantFunctionBody)
					fmt.Fprint(configOutput, wantConfig)
					return nil
				}).
				Times(1)
		},
	}
	testSubcommand(t, args)

	gotFunctionBody, _ := os.ReadFile(importedFunctionPath)
	if diff := cmp.Diff(wantFunctionBody, string(gotFunctionBody)); diff != "" {
		t.Errorf("function body (-want, +got):\n%s", diff)
	}
	gotConfig, _ := os.ReadFile(importedConfigPath)
	if diff := cmp.Diff(wantConfig, string(gotConfig)); diff != "" {
		t.Errorf("config (-want, +got):\n%s", diff)
	}
}

type mockWithLogger[M any] struct {
	M      M
	Logger testLogger
}

type testLogger interface {
	Log(...any)
	Logf(string, ...any)
}

type testSubcommandArgs struct {
	expectDeploy func(m *mockWithLogger[*cli.MockDeployController])
	expectImport func(m *mockWithLogger[*cli.MockImportController])
	expectRender func(m *mockWithLogger[*cli.MockRenderController])
	args         []string
	expect       testSubommandExpectation
}

type testSubommandExpectation struct {
	err error
}

func testSubcommand(t *testing.T, args testSubcommandArgs) {
	t.Helper()

	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	ctrl := gomock.NewController(t)
	deployCtrl := cli.NewMockDeployController(ctrl)
	importCtrl := cli.NewMockImportController(ctrl)
	renderCtrl := cli.NewMockRenderController(ctrl)
	controllers := cli.Controllers{
		DeployController: deployCtrl,
		ImportController: importCtrl,
		RenderController: renderCtrl,
	}
	if args.expectDeploy != nil {
		args.expectDeploy(&mockWithLogger[*cli.MockDeployController]{M: deployCtrl, Logger: t})
	}
	if args.expectImport != nil {
		args.expectImport(&mockWithLogger[*cli.MockImportController]{M: importCtrl, Logger: t})
	}
	if args.expectRender != nil {
		args.expectRender(&mockWithLogger[*cli.MockRenderController]{M: renderCtrl, Logger: t})
	}
	app := cli.New(stdin, stdout, stderr, controllers)

	ctx, cancel := context.WithCancel(context.Background())
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(ctx, deadline)
	}
	defer cancel()
	opts := make([]string, 0, len(args.args)+1)
	opts = append(opts, "frontier")
	opts = append(opts, args.args...)
	t.Logf("args: %s", strings.Join(opts, " "))
	gotErr := app.Run(ctx, opts)
	if diff := diffErrorsConservatively(args.expect.err, gotErr); diff != "" {
		t.Errorf("error (-want, +got):\n%s", diff)
	}
}

type literalError struct {
	msg string
}

func (e *literalError) Error() string { return e.msg }

func (e *literalError) Is(other error) bool {
	if e == nil {
		return other == nil
	}
	if other == nil {
		return false
	}
	return e.msg == other.Error()
}

func diffErrorsConservatively(want, got error) string {
	return cmp.Diff(want, got, cmpopts.EquateErrors())
}

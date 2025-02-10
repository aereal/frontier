package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/aereal/frontier"
	"github.com/aereal/frontier/controller/listdist"
	"github.com/aereal/frontier/internal/cli"
	"github.com/aereal/frontier/internal/fnarn"
	"github.com/aereal/frontier/internal/testexpectations"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	gomock "go.uber.org/mock/gomock"
)

var (
	fnNameDerivedFromConfig      = "test-func"
	fnArnDerivedFromConfig       = "arn:aws:cloudfront::123456789012:function/" + fnNameDerivedFromConfig
	associationDerivedFromConfig = frontier.FunctionAssociation{
		EventType: "viewer-request",
		Distribution: frontier.AssociatedDistribution{
			DomainName: "dist-1.test",
			ID:         "dist-1",
			ARN:        "arn:aws:cloudfront::123456789012:distribution/dist-1",
			IsEnabled:  true,
			IsStaging:  false,
			Status:     "Deployed",
		},
		CacheBehavior: frontier.CacheBehavior{
			IsDefault:      true,
			CachePolicyID:  "default_policy_1",
			TargetOriginID: "origin_1",
		},
		Function: frontier.AssociatedFunction{
			ARN: fnArnDerivedFromConfig,
		},
	}
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
		{
			args: []string{"dist", "list"},
			expectListDistributions: func(m *mockWithLogger[*cli.MockListDistributionsController]) {
				out := []frontier.FunctionAssociation{
					testexpectations.FunctionAssociatedInDefaultCacheBehavior,
				}
				m.M.EXPECT().
					ListDistributions(gomock.Any(), gomock.Any(), listdist.NewCriteria()).
					Return(out, nil).
					Times(1)
			},
		},
		{
			args: []string{"dist", "list", "--event-type", "viewer-request"},
			expectListDistributions: func(m *mockWithLogger[*cli.MockListDistributionsController]) {
				out := []frontier.FunctionAssociation{
					testexpectations.FunctionAssociatedInDefaultCacheBehavior,
				}
				m.M.EXPECT().
					ListDistributions(gomock.Any(), gomock.Any(), listdist.NewCriteria(listdist.EqualEventType("viewer-request"))).
					Return(out, nil).
					Times(1)
			},
		},
		{
			args: []string{"dist", "list", "--function-arn", testexpectations.FunctionArn},
			expectListDistributions: func(m *mockWithLogger[*cli.MockListDistributionsController]) {
				out := []frontier.FunctionAssociation{
					testexpectations.FunctionAssociatedInDefaultCacheBehavior,
				}
				m.M.EXPECT().
					ListDistributions(gomock.Any(), gomock.Any(), listdist.NewCriteria(listdist.EqualFunctionArn(testexpectations.FunctionArn))).
					Return(out, nil).
					Times(1)
			},
		},
		{
			args: []string{"dist", "list", "--function-name", testexpectations.FunctionName},
			expectFunctionARNResolver: func(m *mockWithLogger[*cli.MockFunctionARNResolver]) {
				m.M.EXPECT().
					ResolveFunctionARN(gomock.Any(), fnarn.FunctionName(testexpectations.FunctionName)).
					Return(testexpectations.FunctionArn, nil).
					Times(1)
			},
			expectListDistributions: func(m *mockWithLogger[*cli.MockListDistributionsController]) {
				out := []frontier.FunctionAssociation{
					testexpectations.FunctionAssociatedInDefaultCacheBehavior,
				}
				m.M.EXPECT().
					ListDistributions(gomock.Any(), gomock.Any(), listdist.NewCriteria(listdist.EqualFunctionArn(testexpectations.FunctionArn))).
					Return(out, nil).
					Times(1)
			},
		},
		{
			args: []string{"dist", "list", "--function-name", testexpectations.FunctionName},
			expectFunctionARNResolver: func(m *mockWithLogger[*cli.MockFunctionARNResolver]) {
				m.M.EXPECT().
					ResolveFunctionARN(gomock.Any(), fnarn.FunctionName(testexpectations.FunctionName)).
					Return("", &literalError{"oops"}).
					Times(1)
			},
			expect: testSubommandExpectation{err: &literalError{"oops"}},
		},
		{
			args: []string{"dist", "list", "--config", configPath, "--current"},
			expectFunctionARNResolver: func(m *mockWithLogger[*cli.MockFunctionARNResolver]) {
				m.M.EXPECT().
					ResolveFunctionARN(gomock.Any(), fnarn.FunctionName(fnNameDerivedFromConfig)).
					Return(fnArnDerivedFromConfig, nil).
					Times(1)
			},
			expectListDistributions: func(m *mockWithLogger[*cli.MockListDistributionsController]) {
				out := []frontier.FunctionAssociation{associationDerivedFromConfig}
				m.M.EXPECT().
					ListDistributions(gomock.Any(), gomock.Any(), listdist.NewCriteria(listdist.EqualFunctionArn(fnArnDerivedFromConfig))).
					Return(out, nil).
					Times(1)
			},
		},
		{
			args: []string{"dist", "list", "--config", configPath, "--current"},
			expectFunctionARNResolver: func(m *mockWithLogger[*cli.MockFunctionARNResolver]) {
				m.M.EXPECT().
					ResolveFunctionARN(gomock.Any(), fnarn.FunctionName(fnNameDerivedFromConfig)).
					Return("", &literalError{"oops"}).
					Times(1)
			},
			expect: testSubommandExpectation{err: &literalError{"oops"}},
		},
		{
			args:   []string{"dist", "list", "--config", "not_found.yml", "--current"},
			expect: testSubommandExpectation{&literalError{"os.Open: open not_found.yml: no such file or directory"}},
		},
		{
			args: []string{"dist", "list", "--format", "unknown"},
			expect: testSubommandExpectation{
				err: &literalError{`invalid value "unknown" for flag -format: invalid output format: "unknown"`},
			},
		},
		{
			args: []string{"dist", "list"},
			expectListDistributions: func(m *mockWithLogger[*cli.MockListDistributionsController]) {
				m.M.EXPECT().
					ListDistributions(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, &literalError{"oops"}).
					Times(1)
			},
			expect: testSubommandExpectation{
				err: &literalError{"oops"},
			},
		},
	}
	for idx, tc := range tcs {
		tc := tc
		t.Run(strconv.Itoa(idx)+strings.Join(tc.args, " "), func(t *testing.T) {
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
	expectDeploy              func(m *mockWithLogger[*cli.MockDeployController])
	expectImport              func(m *mockWithLogger[*cli.MockImportController])
	expectRender              func(m *mockWithLogger[*cli.MockRenderController])
	expectListDistributions   func(m *mockWithLogger[*cli.MockListDistributionsController])
	expectFunctionARNResolver func(m *mockWithLogger[*cli.MockFunctionARNResolver])
	args                      []string
	expect                    testSubommandExpectation
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
	listDistsCtrl := cli.NewMockListDistributionsController(ctrl)
	controllers := cli.Controllers{
		DeployController:            deployCtrl,
		ImportController:            importCtrl,
		RenderController:            renderCtrl,
		ListDistributionsController: listDistsCtrl,
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
	if args.expectListDistributions != nil {
		m := &mockWithLogger[*cli.MockListDistributionsController]{M: listDistsCtrl, Logger: t}
		args.expectListDistributions(m)
	}
	arnResolver := cli.NewMockFunctionARNResolver(ctrl)
	if args.expectFunctionARNResolver != nil {
		m := &mockWithLogger[*cli.MockFunctionARNResolver]{
			M:      arnResolver,
			Logger: t,
		}
		args.expectFunctionARNResolver(m)
	}
	app := cli.New(stdin, stdout, stderr, controllers, arnResolver)

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

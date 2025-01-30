package cli_test

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"strings"
	"testing"

	"github.com/aereal/frontier"
	"github.com/aereal/frontier/internal/cli"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/mock/gomock"
)

func TestApp_import(t *testing.T) {
	configPath := "../../testdata/config.yml"
	tcs := []testSubcommandArgs{
		{
			args: []string{"import", "--config", configPath},
			expectImporter: func(m *mockWithLogger[*cli.MockImporter]) {
				m.M.EXPECT().
					Import(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, s string, w io.Writer, wf *frontier.WritableFile) error {
						t.Logf("")
						return nil
					}).
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

func TestApp_deploy(t *testing.T) {
	configPath := "../../testdata/config.yml"
	tcs := []testSubcommandArgs{
		{
			args: []string{"deploy", "--config", configPath},
			expectDeployer: func(m *mockWithLogger[*cli.MockDeployer]) {
				m.M.EXPECT().Deploy(gomock.Any(), configPath, true).Times(1).Return(nil)
			},
		},
		{
			args: []string{"deploy", "--config", configPath, "--publish=false"},
			expectDeployer: func(m *mockWithLogger[*cli.MockDeployer]) {
				m.M.EXPECT().Deploy(gomock.Any(), configPath, false).Times(1).Return(nil)
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

func TestApp_otel(t *testing.T) {
	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	otlpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(otlpServer.Close)
	ctrl := gomock.NewController(t)
	deployer := cli.NewMockDeployer(ctrl)
	importer := cli.NewMockImporter(ctrl)
	renderer := cli.NewMockRenderer(ctrl)
	renderer.EXPECT().
		Render(gomock.Any(), "../../testdata/config.yml", stdout).
		Return(nil).
		Times(1)
	app := cli.New(stdin, stdout, stderr, httpExporterFactory{}, deployer, importer, renderer)

	ctx, cancel := context.WithCancel(context.Background())
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(ctx, deadline)
	}
	defer cancel()
	url, err := neturl.Parse(otlpServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	args := []string{
		"frontier", "--otel-trace-endpoint", net.JoinHostPort(url.Hostname(), url.Port()),
		"render", "--config", "../../testdata/config.yml",
	}
	gotErr := app.Run(ctx, args)
	if gotErr != nil {
		t.Errorf("got error: %T %s", gotErr, gotErr)
	}
}

type httpExporterFactory struct{}

var _ cli.SpanExporterFactory = (*httpExporterFactory)(nil)

func (httpExporterFactory) BuildSpanExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	return otlptracehttp.New(ctx, otlptracehttp.WithInsecure(), otlptracehttp.WithEndpoint(endpoint))
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
	expectDeployer func(m *mockWithLogger[*cli.MockDeployer])
	expectImporter func(m *mockWithLogger[*cli.MockImporter])
	args           []string
}

func testSubcommand(t *testing.T, args testSubcommandArgs) {
	t.Helper()

	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	ctrl := gomock.NewController(t)
	deployer := cli.NewMockDeployer(ctrl)
	importer := cli.NewMockImporter(ctrl)
	renderer := cli.NewMockRenderer(ctrl)
	if args.expectDeployer != nil {
		args.expectDeployer(&mockWithLogger[*cli.MockDeployer]{M: deployer, Logger: t})
	}
	if args.expectImporter != nil {
		args.expectImporter(&mockWithLogger[*cli.MockImporter]{M: importer, Logger: t})
	}
	app := cli.New(stdin, stdout, stderr, cli.NoopExporterFactory{}, deployer, importer, renderer)

	ctx, cancel := context.WithCancel(context.Background())
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(ctx, deadline)
	}
	defer cancel()
	opts := make([]string, 0, len(args.args)+1)
	opts = append(opts, "frontier")
	opts = append(opts, args.args...)
	gotErr := app.Run(ctx, opts)
	if gotErr != nil {
		t.Errorf("got error: %T %s", gotErr, gotErr)
	}
}

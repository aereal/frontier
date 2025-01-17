package cli_test

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"testing"

	"github.com/aereal/frontier/internal/cli"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestApp_otel(t *testing.T) {
	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	otlpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(otlpServer.Close)
	app := cli.New(stdin, stdout, stderr, httpExporterFactory{})

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

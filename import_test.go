package frontier_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/aereal/frontier"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/smithy-go"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
)

func TestImporter_Import(t *testing.T) {
	testCases := []struct {
		name         string
		mock         func(c *MockCFForImport)
		wantConfig   string
		wantFunction string
		wantErr      error
	}{
		{
			name: "ok",
			mock: func(c *MockCFForImport) {
				okGetFunction(c)
				okDescribeFunction(c)
			},
			wantFunction: identityFunction,
			wantConfig:   okFunctionConfig,
		},
		{
			name:    "failed to call GetFunction()",
			wantErr: errOops,
			mock: func(c *MockCFForImport) {
				c.EXPECT().
					GetFunction(gomock.Any(), gomock.Any()).
					Return(nil, errOops).
					Times(1)
			},
		},
		{
			name:    "failed to call DescribeFunction()",
			wantErr: errOops,
			mock: func(c *MockCFForImport) {
				okGetFunction(c)
				c.EXPECT().
					DescribeFunction(gomock.Any(), gomock.Any()).
					Return(nil, errOops).
					Times(1)
			},
		},
		{
			name:    "function not found",
			wantErr: errNoSuchFn,
			mock: func(c *MockCFForImport) {
				err := &smithy.OperationError{
					Err: errNoSuchFn,
				}
				c.EXPECT().
					GetFunction(gomock.Any(), gomock.Any()).
					Return(nil, err).
					Times(1)
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			if deadline, ok := t.Deadline(); ok {
				ctx, cancel = context.WithDeadline(ctx, deadline)
			}
			defer cancel()

			ctrl := gomock.NewController(t)
			client := NewMockCFForImport(ctrl)
			if tc.mock != nil {
				tc.mock(client)
			}
			configOut := new(bytes.Buffer)
			fnOut := new(bytes.Buffer)
			wf := &frontier.WritableFile{
				FilePath: "test-fn.js",
				Writer:   fnOut,
			}
			importer := frontier.NewImporter(client, "test-fn", configOut, wf)
			gotErr := importer.Import(ctx)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("error:\n\twant: %s (%T)\n\t got: %s (%T)", tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			gotConfig := configOut.String()
			if diff := cmp.Diff(tc.wantConfig, gotConfig); diff != "" {
				t.Errorf("config file (-want, +got):\n%s", diff)
			}
			gotFunction := fnOut.String()
			if diff := cmp.Diff(tc.wantFunction, gotFunction); diff != "" {
				t.Errorf("function file (-want, +got):\n%s", diff)
			}
		})
	}
}

var (
	errOops     = errors.New("oops")
	errNoSuchFn = &types.NoSuchFunctionExists{
		Message: ref("function not found"),
	}
	okFunctionConfig = `name: test-fn
code:
  path: test-fn.js
config:
  comment: blah blah
  runtime: cloudfront-js-2.0
`
	identityFunction = `
function handler(event) { return event.response }
`
	okGetFunctionOutput = &cloudfront.GetFunctionOutput{
		ContentType:  ref("application/octet-stream"),
		ETag:         ref("0xdeadbeaf"),
		FunctionCode: []byte(identityFunction),
	}
	okDescribeFunctionOutput = &cloudfront.DescribeFunctionOutput{
		ETag: ref("0xdeadbeaf"),
		FunctionSummary: &types.FunctionSummary{
			Name:   ref("test-fn"),
			Status: ref("DEPLOYED"),
			FunctionConfig: &types.FunctionConfig{
				Comment: ref("blah blah"),
				Runtime: types.FunctionRuntimeCloudfrontJs20,
			},
			FunctionMetadata: &types.FunctionMetadata{
				FunctionARN: ref("arn:aws:cloudfront::123456789012:function/test-fn"),
				Stage:       types.FunctionStageLive,
			},
		},
	}
)

func okGetFunction(c *MockCFForImport) {
	c.EXPECT().
		GetFunction(gomock.Any(), gomock.Any()).
		Return(okGetFunctionOutput, nil).
		Times(1)
}

func okDescribeFunction(c *MockCFForImport) {
	c.EXPECT().
		DescribeFunction(gomock.Any(), gomock.Any()).
		Return(okDescribeFunctionOutput, nil).
		Times(1)
}

package fnarn_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aereal/frontier/internal/cf"
	"github.com/aereal/frontier/internal/cfmock"
	"github.com/aereal/frontier/internal/fnarn"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
)

const (
	functionArn = "arn:aws:cloudfront::123456789012:function/test-fn"
)

func TestResolver_ResolveFunctionARN(t *testing.T) {
	testCases := []struct {
		name         string
		identifier   fnarn.FunctionIdentifier
		wantArn      string
		wantErr      error
		expectClient func(m *cfmock.MockCloudFrontClient)
	}{
		{
			name:       "arn",
			identifier: fnarn.FunctionARN(functionArn),
			wantArn:    functionArn,
		},
		{
			name:       "pass name/DescribeFunction success",
			identifier: fnarn.FunctionName("test-fn"),
			wantArn:    functionArn,
			expectClient: func(m *cfmock.MockCloudFrontClient) {
				out := &cloudfront.DescribeFunctionOutput{
					FunctionSummary: &types.FunctionSummary{
						FunctionMetadata: &types.FunctionMetadata{
							FunctionARN: ref(functionArn),
						},
					},
				}
				m.EXPECT().
					DescribeFunction(gomock.Any(), gomock.Any()).
					Return(out, nil).
					Times(1)
			},
		},
		{
			name:       "pass name/DescribeFunction FAILED",
			identifier: fnarn.FunctionName("test-fn"),
			wantErr:    sdkError{},
			expectClient: func(m *cfmock.MockCloudFrontClient) {
				m.EXPECT().
					DescribeFunction(gomock.Any(), gomock.Any()).
					Return(nil, sdkError{}).
					Times(1)
			},
		},
		{
			name:       "pass unknown identifier",
			identifier: nil,
			wantErr:    &fnarn.UnsupportedFunctionIdentifierError{T: reflect.TypeOf(nil)},
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
			client := cfmock.NewMockCloudFrontClient(ctrl)
			if tc.expectClient != nil {
				tc.expectClient(client)
			}
			resolver := fnarn.NewResolver(&cf.StaticCFProvider{Client: client})
			gotArn, gotErr := resolver.ResolveFunctionARN(ctx, tc.identifier)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("error:\n\twant: <%T> %s\n\t got: <%T> %s", tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if diff := cmp.Diff(tc.wantArn, gotArn); diff != "" {
				t.Errorf("arn (-want, +got):\n%s", diff)
			}
		})
	}
}

func ref[T any](v T) *T { return &v }

type sdkError struct{}

func (sdkError) Error() string { return "error from AWS SDK" }

package frontier_test

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/aereal/frontier"
	"github.com/aereal/frontier/internal/cfmock"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"
)

//go:embed testdata/fn.js
var functionCode []byte

func TestDeployer_ok_existing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(ctx, deadline)
	}
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := cfmock.NewMockCloudFrontClient(ctrl)
	client.EXPECT().
		GetFunction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *cloudfront.GetFunctionInput, _ ...func(*cloudfront.Options)) (*cloudfront.GetFunctionOutput, error) {
			want := &cloudfront.GetFunctionInput{Name: ref("test-func")}
			if diff := compare(want, input); diff != "" {
				t.Errorf("GetFunctionInput (-want, +got)\n:%s", diff)
			}
			return &cloudfront.GetFunctionOutput{
				ContentType: ref("application/json"),
				ETag:        ref("0xdeadbeaf"),
			}, nil
		}).
		Times(1)
	client.EXPECT().
		UpdateFunction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *cloudfront.UpdateFunctionInput, _ ...func(*cloudfront.Options)) (*cloudfront.UpdateFunctionOutput, error) {
			want := &cloudfront.UpdateFunctionInput{
				FunctionCode: functionCode,
				FunctionConfig: &types.FunctionConfig{
					Comment: ref("blah blah"),
					Runtime: types.FunctionRuntimeCloudfrontJs10,
				},
				IfMatch: ref("0xdeadbeaf"),
				Name:    ref("test-func"),
			}
			if diff := compare(want, input); diff != "" {
				t.Errorf("UpdateFunctionInput (-want, +got):\n%s", diff)
			}
			return &cloudfront.UpdateFunctionOutput{
				ETag: ref("updated-0xdeadbeaf"),
			}, nil
		}).
		Times(1)
	client.EXPECT().
		PublishFunction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *cloudfront.PublishFunctionInput, _ ...func(*cloudfront.Options)) (*cloudfront.PublishFunctionOutput, error) {
			now := time.Now()
			return &cloudfront.PublishFunctionOutput{
				FunctionSummary: &types.FunctionSummary{
					Name: ref("test-func"),
					FunctionConfig: &types.FunctionConfig{
						Comment: ref("blah blah"),
						Runtime: types.FunctionRuntimeCloudfrontJs10,
					},
					FunctionMetadata: &types.FunctionMetadata{
						CreatedTime:      ref(now),
						LastModifiedTime: ref(now),
						Stage:            types.FunctionStageLive,
					},
				},
			}, nil
		}).
		Times(1)

	deployer := frontier.NewDeployer(client)
	if err := deployer.Deploy(ctx, "./testdata/config.yml", true); err != nil {
		t.Errorf("deployer.Deploy: %+v", err)
	}
}

func TestDeployer_ok_create(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(ctx, deadline)
	}
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := cfmock.NewMockCloudFrontClient(ctrl)
	client.EXPECT().
		GetFunction(gomock.Any(), gomock.Any()).
		Return(nil, &types.NoSuchFunctionExists{Message: ref("not found")}).
		Times(1)
	client.EXPECT().
		CreateFunction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *cloudfront.CreateFunctionInput, _ ...func(*cloudfront.Options)) (*cloudfront.CreateFunctionOutput, error) {
			want := &cloudfront.CreateFunctionInput{
				FunctionCode: functionCode,
				FunctionConfig: &types.FunctionConfig{
					Comment: ref("blah blah"),
					Runtime: types.FunctionRuntimeCloudfrontJs10,
				},
				Name: ref("test-func"),
			}
			if diff := compare(want, input); diff != "" {
				t.Errorf("CreateFunctionInput (-want, +got):\n%s", diff)
			}
			return &cloudfront.CreateFunctionOutput{
				ETag: ref("created-0xdeadbeaf"),
			}, nil
		}).
		Times(1)
	client.EXPECT().
		PublishFunction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *cloudfront.PublishFunctionInput, _ ...func(*cloudfront.Options)) (*cloudfront.PublishFunctionOutput, error) {
			now := time.Now()
			return &cloudfront.PublishFunctionOutput{
				FunctionSummary: &types.FunctionSummary{
					Name: ref("test-func"),
					FunctionConfig: &types.FunctionConfig{
						Comment: ref("blah blah"),
						Runtime: types.FunctionRuntimeCloudfrontJs10,
					},
					FunctionMetadata: &types.FunctionMetadata{
						CreatedTime:      ref(now),
						LastModifiedTime: ref(now),
						Stage:            types.FunctionStageLive,
					},
				},
			}, nil
		}).
		Times(1)

	deployer := frontier.NewDeployer(client)
	if err := deployer.Deploy(ctx, "./testdata/config.yml", true); err != nil {
		t.Errorf("deployer.Deploy: %+v", err)
	}
}

func ref[T any](v T) *T {
	return &v
}

func compare(want, got any) string {
	return cmp.Diff(want, got, cmpopts.IgnoreUnexported(cloudfront.GetFunctionInput{}, cloudfront.UpdateFunctionInput{}, cloudfront.CreateFunctionInput{}, types.FunctionConfig{}))
}

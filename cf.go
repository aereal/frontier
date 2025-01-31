package frontier

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

type CloudFrontClient interface {
	CFForDeploy
	CFForImport
}

type CFForDeploy interface {
	CreateFunction(ctx context.Context, params *cloudfront.CreateFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.CreateFunctionOutput, error)
	GetFunction(ctx context.Context, params *cloudfront.GetFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.GetFunctionOutput, error)
	PublishFunction(ctx context.Context, params *cloudfront.PublishFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.PublishFunctionOutput, error)
	UpdateFunction(ctx context.Context, params *cloudfront.UpdateFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.UpdateFunctionOutput, error)
}

type CFForImport interface {
	DescribeFunction(ctx context.Context, params *cloudfront.DescribeFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.DescribeFunctionOutput, error)
	GetFunction(ctx context.Context, params *cloudfront.GetFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.GetFunctionOutput, error)
}

type CloudFrontClientProvider interface {
	ProvideCloudFrontClient(ctx context.Context) (CloudFrontClient, error)
}

type StaticCFClientProvider struct {
	Client CloudFrontClient
}

var _ CloudFrontClientProvider = (*StaticCFClientProvider)(nil)

func (s *StaticCFClientProvider) ProvideCloudFrontClient(_ context.Context) (CloudFrontClient, error) {
	return s.Client, nil
}

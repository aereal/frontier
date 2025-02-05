package cf

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

type CloudFrontClient interface {
	CreateFunction(ctx context.Context, params *cloudfront.CreateFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.CreateFunctionOutput, error)
	DescribeFunction(ctx context.Context, params *cloudfront.DescribeFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.DescribeFunctionOutput, error)
	GetFunction(ctx context.Context, params *cloudfront.GetFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.GetFunctionOutput, error)
	PublishFunction(ctx context.Context, params *cloudfront.PublishFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.PublishFunctionOutput, error)
	UpdateFunction(ctx context.Context, params *cloudfront.UpdateFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.UpdateFunctionOutput, error)
}

type Provider interface {
	ProvideCloudFrontClient(ctx context.Context) (CloudFrontClient, error)
}

type StaticCFProvider struct {
	Client CloudFrontClient
}

var _ Provider = (*StaticCFProvider)(nil)

func (p *StaticCFProvider) ProvideCloudFrontClient(context.Context) (CloudFrontClient, error) { //nolint:ireturn
	return p.Client, nil
}

type SDKProvider struct{}

var _ Provider = (*SDKProvider)(nil)

func (b SDKProvider) ProvideCloudFrontClient(ctx context.Context) (CloudFrontClient, error) { //nolint:ireturn
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	otelaws.AppendMiddlewares(&cfg.APIOptions)
	return cloudfront.NewFromConfig(cfg), nil
}

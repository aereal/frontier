package cli

import (
	"context"

	"github.com/aereal/frontier"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

type CloudFrontSDKBuilder struct{}

var _ frontier.CloudFrontClientProvider = (*CloudFrontSDKBuilder)(nil)

func (b CloudFrontSDKBuilder) ProvideCloudFrontClient(ctx context.Context) (frontier.CloudFrontClient, error) { //nolint:ireturn
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	otelaws.AppendMiddlewares(&cfg.APIOptions)
	return cloudfront.NewFromConfig(cfg), nil
}

//go:generate go run go.uber.org/mock/mockgen -typed -write_command_comment=false -write_package_comment=false -write_source_comment=false -package frontier_test -destination mock_cloudfront_client_test.go github.com/aereal/frontier CloudFrontClient

package frontier

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

type CloudFrontClient interface {
	CreateFunction(ctx context.Context, params *cloudfront.CreateFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.CreateFunctionOutput, error)
	GetFunction(ctx context.Context, params *cloudfront.GetFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.GetFunctionOutput, error)
	PublishFunction(ctx context.Context, params *cloudfront.PublishFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.PublishFunctionOutput, error)
	UpdateFunction(ctx context.Context, params *cloudfront.UpdateFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.UpdateFunctionOutput, error)
}

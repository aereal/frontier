package fnarn

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/aereal/frontier/internal/cf"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

func NewResolver(clientProvider cf.Provider) *Resolver {
	return &Resolver{clientProvider: clientProvider}
}

type Resolver struct {
	clientProvider cf.Provider
}

func (r *Resolver) ResolveFunctionARN(ctx context.Context, identifier FunctionIdentifier) (string, error) {
	if arn, ok := identifier.(FunctionARN); ok {
		return string(arn), nil
	}
	if name, ok := identifier.(FunctionName); ok {
		s := string(name)
		input := &cloudfront.DescribeFunctionInput{
			Name: &s,
		}
		client, err := r.clientProvider.ProvideCloudFrontClient(ctx)
		if err != nil {
			return "", err
		}
		out, err := client.DescribeFunction(ctx, input)
		if err != nil {
			return "", err
		}
		if out.FunctionSummary != nil && out.FunctionSummary.FunctionMetadata != nil && out.FunctionSummary.FunctionMetadata.FunctionARN != nil {
			return *out.FunctionSummary.FunctionMetadata.FunctionARN, nil
		}
	}
	return "", &UnsupportedFunctionIdentifierError{T: reflect.TypeOf(identifier)}
}

type FunctionIdentifier interface {
	isFunctionIdentifier()
}

type FunctionName string

var _ FunctionIdentifier = (FunctionName)("")

func (FunctionName) isFunctionIdentifier() {}

type FunctionARN string

var _ FunctionIdentifier = (FunctionARN)("")

func (FunctionARN) isFunctionIdentifier() {}

type UnsupportedFunctionIdentifierError struct {
	T reflect.Type
}

func (e *UnsupportedFunctionIdentifierError) Error() string {
	return fmt.Sprintf("unsupported function identifier type: %s", e.T)
}

func (e *UnsupportedFunctionIdentifierError) Is(other error) bool {
	if e == nil {
		return other == nil
	}
	err := new(UnsupportedFunctionIdentifierError)
	if errors.As(other, &err) {
		return e.T == err.T
	}
	return false
}

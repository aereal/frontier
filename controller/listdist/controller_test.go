package listdist_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/aereal/frontier"
	"github.com/aereal/frontier/controller/listdist"
	"github.com/aereal/frontier/internal/cf"
	"github.com/aereal/frontier/internal/cfmock"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
)

func TestController_ListDistributions(t *testing.T) {
	testCases := []struct {
		name             string
		criteria         *listdist.Criteria
		expectClient     func(m *cfmock.MockCloudFrontClient)
		wantAssociations []frontier.FunctionAssociation
		wantErr          error
	}{
		{
			name:             "returned distriution that associated function in default cache behavior",
			criteria:         listdist.NewCriteria(),
			expectClient:     returnDist(distAssociatedInDefaultCacheBehavior),
			wantAssociations: []frontier.FunctionAssociation{want1},
		},
		{
			name:             "returned distriution that associated function in custom cache behavior",
			criteria:         listdist.NewCriteria(),
			expectClient:     returnDist(distAssociatedInCustomCacheBehavior),
			wantAssociations: []frontier.FunctionAssociation{want2},
		},
		{
			name:         "AWS returned error",
			criteria:     listdist.NewCriteria(),
			expectClient: returnApiError(),
			wantErr:      apiError{},
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

			mockCtrl := gomock.NewController(t)
			client := cfmock.NewMockCloudFrontClient(mockCtrl)
			if tc.expectClient != nil {
				tc.expectClient(client)
			}
			controller := listdist.NewController(&cf.StaticCFProvider{Client: client})
			buf := new(bytes.Buffer)
			gotAssociations, gotErr := controller.ListDistributions(ctx, buf, tc.criteria)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("error:\n\twant: %T %s\n\t got: %T %s", tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if gotErr != nil {
				return
			}
			if diff := cmp.Diff(tc.wantAssociations, gotAssociations); diff != "" {
				t.Errorf("associations (-want, +got):\n%s", diff)
			}
		})
	}
}

func returnApiError() func(m *cfmock.MockCloudFrontClient) {
	return func(m *cfmock.MockCloudFrontClient) {
		m.EXPECT().
			ListDistributions(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, apiError{}).
			Times(1)
	}
}

func returnDist(dist types.DistributionSummary) func(m *cfmock.MockCloudFrontClient) {
	return func(m *cfmock.MockCloudFrontClient) {
		out := &cloudfront.ListDistributionsOutput{
			DistributionList: &types.DistributionList{
				Items: []types.DistributionSummary{dist},
			},
		}
		m.EXPECT().
			ListDistributions(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(out, nil).
			Times(1)
	}
}

var (
	functionArn = "arn:aws:cloudfront::123456789012:function/test-fn"

	want1 = frontier.FunctionAssociation{
		EventType: "viewer-request",
		Distribution: frontier.AssociatedDistribution{
			DomainName: "dist-1.test",
			ID:         "dist-1",
			ARN:        "arn:aws:cloudfront::123456789012:distribution/dist-1",
			IsEnabled:  true,
			IsStaging:  false,
			Status:     "Deployed",
		},
		CacheBehavior: frontier.CacheBehavior{
			IsDefault:      true,
			CachePolicyID:  "default_policy_1",
			TargetOriginID: "origin_1",
		},
		Function: frontier.AssociatedFunction{
			ARN: functionArn,
		},
	}
	want2 = frontier.FunctionAssociation{
		EventType: "viewer-request",
		Distribution: frontier.AssociatedDistribution{
			DomainName: "dist-2.test",
			ID:         "dist-2",
			ARN:        "arn:aws:cloudfront::123456789012:distribution/dist-2",
			IsEnabled:  true,
			IsStaging:  false,
			Status:     "Deployed",
		},
		CacheBehavior: frontier.CacheBehavior{
			IsDefault:      false,
			CachePolicyID:  "policy_1",
			TargetOriginID: "origin_2",
		},
		Function: frontier.AssociatedFunction{
			ARN: functionArn,
		},
	}
	distAssociatedInDefaultCacheBehavior = types.DistributionSummary{
		Id:             ref("dist-1"),
		ARN:            ref("arn:aws:cloudfront::123456789012:distribution/dist-1"),
		DomainName:     ref("dist-1.test"),
		Enabled:        ref(true),
		Staging:        ref(false),
		Status:         ref("Deployed"),
		CacheBehaviors: &types.CacheBehaviors{},
		DefaultCacheBehavior: &types.DefaultCacheBehavior{
			TargetOriginId: ref("origin_1"),
			CachePolicyId:  ref("default_policy_1"),
			FunctionAssociations: &types.FunctionAssociations{
				Items: []types.FunctionAssociation{
					{FunctionARN: &functionArn, EventType: "viewer-request"},
				},
			},
		},
	}
	distAssociatedInCustomCacheBehavior = types.DistributionSummary{
		Id:         ref("dist-2"),
		ARN:        ref("arn:aws:cloudfront::123456789012:distribution/dist-2"),
		DomainName: ref("dist-2.test"),
		Enabled:    ref(true),
		Staging:    ref(false),
		Status:     ref("Deployed"),
		CacheBehaviors: &types.CacheBehaviors{
			Items: []types.CacheBehavior{
				{
					TargetOriginId: ref("origin_2"),
					CachePolicyId:  ref("policy_1"),
					FunctionAssociations: &types.FunctionAssociations{
						Items: []types.FunctionAssociation{
							{FunctionARN: &functionArn, EventType: "viewer-request"},
						},
					},
				},
			},
		},
		DefaultCacheBehavior: &types.DefaultCacheBehavior{
			TargetOriginId:       ref("origin_1"),
			CachePolicyId:        ref("default_policy_1"),
			FunctionAssociations: &types.FunctionAssociations{},
		},
	}
)

func ref[T any](v T) *T { return &v }

type apiError struct{}

func (apiError) Error() string { return "api error" }

func (apiError) Is(other error) bool {
	var x apiError
	return errors.As(other, &x)
}

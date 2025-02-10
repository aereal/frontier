package testexpectations

import (
	"github.com/aereal/frontier"
)

var (
	FunctionName = "test-fn"
	FunctionArn  = "arn:aws:cloudfront::123456789012:function/" + FunctionName

	FunctionAssociatedInDefaultCacheBehavior = frontier.FunctionAssociation{
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
			ARN: FunctionArn,
		},
	}
	FunctionAssociatedInCustomCacheBehavior = frontier.FunctionAssociation{
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
			ARN: FunctionArn,
		},
	}
)

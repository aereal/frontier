package presenter

import "github.com/aereal/frontier"

type AssociatedDistributionsPresenter interface {
	PresentAssociatedDistributions(associations []frontier.FunctionAssociation)
}

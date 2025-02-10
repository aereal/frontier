package json

import (
	"encoding/json"
	"io"

	"github.com/aereal/frontier"
	"github.com/aereal/frontier/internal/presenter"
)

type NewAssociatedDistributionsPresenterOption interface {
	applyNewAssociatedDistributionsPresenterOption(cfg *configNewAssociatedDistributionsPresenter)
}

type configNewAssociatedDistributionsPresenter struct {
	pretty bool
}

func Pretty(pretty bool) NewAssociatedDistributionsPresenterOption { return &optPretty{pretty: pretty} } //nolint:ireturn

type optPretty struct{ pretty bool }

var (
	_ NewAssociatedDistributionsPresenterOption = (*optPretty)(nil)
)

func (o *optPretty) applyNewAssociatedDistributionsPresenterOption(cfg *configNewAssociatedDistributionsPresenter) {
	cfg.pretty = o.pretty
}

func NewAssociatedDistributionsPresenter(out io.Writer, opts ...NewAssociatedDistributionsPresenterOption) *AssociatedDistributionsPresenter {
	var cfg configNewAssociatedDistributionsPresenter
	for _, o := range opts {
		o.applyNewAssociatedDistributionsPresenterOption(&cfg)
	}
	enc := json.NewEncoder(out)
	if cfg.pretty {
		enc.SetIndent("", "  ")
	}
	return &AssociatedDistributionsPresenter{
		enc: enc,
	}
}

type AssociatedDistributionsPresenter struct {
	enc *json.Encoder
}

var _ presenter.AssociatedDistributionsPresenter = (*AssociatedDistributionsPresenter)(nil)

func (p *AssociatedDistributionsPresenter) PresentAssociatedDistributions(associations []frontier.FunctionAssociation) {
	for _, a := range associations {
		_ = p.enc.Encode(a) //nolint:errchkjson
	}
}

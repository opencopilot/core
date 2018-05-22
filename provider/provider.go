package provider

import (
	"errors"

	pb "github.com/opencopilot/core/core"
)

// Provider is an instance provider (such as Packet)
type Provider struct {
	PbProvider pb.Provider
}

func (p *Provider) String() (string, error) {
	return p.PbProvider.String(), nil
}

// NewProvider returns a provider
func NewProvider(providerName string) (*Provider, error) {

	p, ok := pb.Provider_value[providerName]
	if !ok {
		return nil, errors.New("Invalid providerName")
	}

	provider := &Provider{
		PbProvider: pb.Provider(p),
	}
	return provider, nil
}

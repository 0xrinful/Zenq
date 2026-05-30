package azora

import (
	"github.com/0xrinful/Zenq/internal/requester"
	"github.com/0xrinful/Zenq/internal/sources"
)

const (
	baseURL    = "https://azoramoon.com"
	apiURL     = "https://api.azoramoon.com"
	sourceID   = "azora"
	sourceName = "Azora"
)

func Config() sources.Config {
	return sources.Config{
		BaseURL:    baseURL,
		NeedsFlare: false,
	}
}

type Source struct {
	req *requester.Requester
}

func New(req *requester.Requester) sources.Source {
	return &Source{req: req}
}

func (s *Source) Info() sources.SourceInfo {
	return sources.SourceInfo{
		ID:   sourceID,
		Name: sourceName,
	}
}

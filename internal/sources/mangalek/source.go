package mangalek

import (
	"github.com/0xrinful/Zenq/internal/requester"
	"github.com/0xrinful/Zenq/internal/sources"
)

const (
	baseURL    = "https://lek-manga.net"
	sourceID   = "mangalek"
	sourceName = "MangaLek"
)

func Config() sources.Config {
	return sources.Config{
		BaseURL:           baseURL,
		CloudflareTestURL: baseURL + "/page/5/",
		NeedsFlare:        true,
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

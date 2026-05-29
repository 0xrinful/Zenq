package registry

import (
	"github.com/0xrinful/Zenq/internal/downloader"
	"github.com/0xrinful/Zenq/internal/requester"
	"github.com/0xrinful/Zenq/internal/requester/flare"
	"github.com/0xrinful/Zenq/internal/sources"
	"github.com/0xrinful/Zenq/internal/sources/mangalek"
)

type Registry struct {
	sources     map[string]sources.Source
	downloaders map[string]*downloader.Downloader
}

func NewRegistry(solver *flare.Solver) *Registry {
	r := &Registry{
		sources:     make(map[string]sources.Source),
		downloaders: make(map[string]*downloader.Downloader),
	}

	r.register(solver, mangalek.Config(), mangalek.New)

	return r
}

func (r *Registry) register(
	solver *flare.Solver,
	cfg sources.Config,
	newSource func(*requester.Requester) sources.Source,
) {
	req := requester.New(solver, cfg)
	source := newSource(req)
	dl := downloader.New(req)

	id := source.Info().ID
	r.sources[id] = source
	r.downloaders[id] = dl
}

func (r *Registry) Source(id string) (sources.Source, bool) {
	s, ok := r.sources[id]
	return s, ok
}

func (r *Registry) Downloader(id string) (*downloader.Downloader, bool) {
	d, ok := r.downloaders[id]
	return d, ok
}

func (r *Registry) AllSources() []sources.Source {
	sources := make([]sources.Source, 0, len(r.sources))
	for _, s := range r.sources {
		sources = append(sources, s)
	}
	return sources
}

func (r *Registry) Sources() []sources.SourceInfo {
	infos := make([]sources.SourceInfo, 0, len(r.sources))
	for _, s := range r.sources {
		infos = append(infos, s.Info())
	}
	return infos
}

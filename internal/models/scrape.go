package models

import "time"

type Manga struct {
	Slug        string
	SourceID    string
	Title       string
	Description string
	CoverURL    string
	Status      string
	Genres      []string
	Chapters    []Chapter
}

type Chapter struct {
	URL        string
	MangaSlug  string
	SourceID   string
	Number     float64
	Title      string
	ReleasedAt time.Time
	Pages      []Page
}

type ChapterRange struct {
	From  float64
	To    float64
	All   bool
	Force bool
}

func (r ChapterRange) Contains(number float64) bool {
	return r.All || (number >= r.From && number <= r.To)
}

type Page struct {
	Number int
	URL    string
}

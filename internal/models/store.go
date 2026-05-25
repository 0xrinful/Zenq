package models

import "time"

type MangaRecord struct {
	Manga
	CoverPath string
	AddedAt   time.Time
	UpdatedAt time.Time
}

type ChapterRecord struct {
	Chapter
	RawPath       string
	OptimizedPath string
	CBZPath       string
	Downloaded    bool
	DownloadedAt  time.Time
	Optimized     bool
	OptimizedAt   time.Time
	Packed        bool
	PackedAt      time.Time
}

type User struct {
	ID           int
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

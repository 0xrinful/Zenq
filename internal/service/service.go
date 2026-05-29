package service

import (
	"errors"

	"github.com/0xrinful/Zenq/internal/packer"
	"github.com/0xrinful/Zenq/internal/queue"
	"github.com/0xrinful/Zenq/internal/registry"
	"github.com/0xrinful/Zenq/internal/storage/db"
	"github.com/0xrinful/Zenq/internal/storage/files"
)

var (
	ErrInvalidCredentials = errors.New("service: invalid credentials")
	ErrUnknownSource      = errors.New("service: unknown source")
	ErrNotFound           = errors.New("service: not found")
	ErrNotDownloaded      = errors.New("service: chapter not downloaded")
)

type Service struct {
	registry *registry.Registry
	db       *db.DB
	files    *files.Store
	queue    *queue.Queue
	packer   *packer.Packer
}

func New(
	registry *registry.Registry,
	db *db.DB,
	files *files.Store,
	queue *queue.Queue,
) *Service {
	return &Service{
		registry: registry,
		db:       db,
		files:    files,
		queue:    queue,
		packer:   packer.New(),
	}
}

func (s *Service) Files() *files.Store {
	return s.files
}

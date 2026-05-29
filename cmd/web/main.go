package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/0xrinful/Zenq/internal/queue"
	"github.com/0xrinful/Zenq/internal/registry"
	"github.com/0xrinful/Zenq/internal/requester/flare"
	"github.com/0xrinful/Zenq/internal/server"
	"github.com/0xrinful/Zenq/internal/service"
	"github.com/0xrinful/Zenq/internal/storage/db"
	"github.com/0xrinful/Zenq/internal/storage/files"
)

func main() {
	secret := os.Getenv("SECRET")
	if secret == "" {
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			log.Fatal(err)
		}
		secret = base64.StdEncoding.EncodeToString(buf)
		slog.Warn("using generated secret for dev")
	}
	server.SetSessionSecret(secret)

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "zenq.db"
	}

	filesRoot := os.Getenv("FILES_ROOT")
	if filesRoot == "" {
		filesRoot = "data"
	}
	if err := os.MkdirAll(filesRoot, 0o755); err != nil {
		log.Fatal(err)
	}

	database, err := db.New(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	fileStore := files.New(filesRoot)
	solver := flare.NewSolver()
	registry := registry.NewRegistry(solver)
	queue := queue.NewQueue()
	svc := service.New(registry, database, fileStore, queue)

	srv := server.New(svc)
	handler := server.WithLogging(srv)

	log.Fatal(http.ListenAndServe(":8080", handler))
}

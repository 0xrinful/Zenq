package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"

	"github.com/0xrinful/Zenq/internal/queue"
	"github.com/0xrinful/Zenq/internal/registry"
	"github.com/0xrinful/Zenq/internal/requester/flare"
	"github.com/0xrinful/Zenq/internal/server"
	"github.com/0xrinful/Zenq/internal/service"
	"github.com/0xrinful/Zenq/internal/storage/db"
	"github.com/0xrinful/Zenq/internal/storage/files"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
	}))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found")
	}

	tlsCertPath := os.Getenv("TLS_CERT_PATH")
	tlsKeyPath := os.Getenv("TLS_KEY_PATH")
	if tlsCertPath == "" || tlsKeyPath == "" {
		log.Fatal("missing TLS certificate path")
	}

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
	} else {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			log.Fatal(err)
		}
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
	q := queue.NewQueue()
	svc := service.New(registry, database, fileStore, q)

	worker := queue.NewWorker(q, registry, queue.Config{
		AutoOptimize: true,
		AutoPack:     true,
	}, database, fileStore)
	worker.Start(context.Background())

	srv := server.New(svc)
	handler := server.WithLogging(srv)

	slog.Info("server started", "addr", 8000)
	log.Fatal(http.ListenAndServeTLS(":8000", tlsCertPath, tlsKeyPath, handler))
}

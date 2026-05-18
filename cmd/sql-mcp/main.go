package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/auth"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/config"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/obs"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/schemas"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/server"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

//go:embed all:embedded
var embedded embed.FS

func embeddedSchemaFiles() (map[string][]byte, error) {
	out := map[string][]byte{}
	entries, err := embedded.ReadDir("embedded")
	if err != nil {
		return nil, fmt.Errorf("embed read: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := embedded.ReadFile("embedded/" + e.Name())
		if err != nil {
			return nil, err
		}
		out[e.Name()] = b
	}
	return out, nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}
	log := obs.NewLogger(cfg.LogLevel, cfg.ServiceName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTracing, err := obs.SetupTracing(ctx, cfg.OTLPEndpoint, cfg.ServiceName)
	if err != nil {
		log.Error("tracing setup failed", "err", err)
		os.Exit(1)
	}
	defer func() { _ = shutdownTracing(context.Background()) }()

	files, err := embeddedSchemaFiles()
	if err != nil {
		log.Error("schema embed read failed", "err", err)
		os.Exit(1)
	}
	cat, err := schemas.LoadFromBytes(files)
	if err != nil {
		log.Error("schema load failed", "err", err)
		os.Exit(1)
	}
	validators, err := schemas.CompileValidators(cat)
	if err != nil {
		log.Error("schema compile failed", "err", err)
		os.Exit(1)
	}
	obs.SchemasLoaded.Set(float64(len(cat.Tools)))
	log.Info("schemas loaded", "count", len(cat.Tools), "digest", cat.Digest)

	pool, err := store.New(ctx, cfg.DatabaseURL, cfg.DBPoolMax, cfg.QueryTimeout())
	if err != nil {
		log.Error("pg pool init failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	jwksURL, err := auth.DiscoverJWKS(ctx, cfg.OIDCIssuer)
	if err != nil {
		log.Error("oidc discovery failed", "issuer", cfg.OIDCIssuer, "err", err)
		os.Exit(1)
	}
	log.Info("oidc discovery ok", "issuer", cfg.OIDCIssuer, "jwks_uri", jwksURL)
	jwt, err := auth.NewValidator(ctx, cfg.OIDCIssuer, cfg.OIDCAudience, jwksURL, cfg.JWKSRefresh())
	if err != nil {
		log.Error("jwt validator init failed", "err", err)
		os.Exit(1)
	}

	deps := server.Deps{
		Log:            log,
		Pool:           pool,
		JWT:            jwt,
		Validators:     validators,
		Catalog:        cat,
		SchemaVersion:  cfg.SchemaVersion,
		ExpectedSPIFFE: cfg.GatewaySPIFFE,
		QueryTimeout:   func() int { return cfg.QueryTimeoutSeconds },
	}
	r := server.NewRouter(deps)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Info("listening", "addr", cfg.ListenAddr, "schema_digest", cat.Digest)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("listen failed", "err", err)
			cancel()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-sigCh:
		log.Info("shutdown signal", "sig", sig.String())
	case <-ctx.Done():
		log.Info("shutdown due to internal error")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown error", "err", err)
	}
}

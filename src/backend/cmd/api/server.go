package main

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), logLevel(app.config.logLevel)),
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	app.logger.Info("starting server", "addr", srv.Addr)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}

func logLevel(s string) slog.Level {
	m := map[string]slog.Level{
		"trace":   slog.Level(-8),
		"debug":   slog.LevelDebug,
		"info":    slog.LevelInfo,
		"warning": slog.LevelWarn,
		"error":   slog.LevelError,
	}

	if level, ok := m[s]; ok {
		return level
	}

	return slog.LevelError
}

//go:embed meals.sqlite
var embeddedDatabaseBytes []byte

func hydrateDB(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	return os.WriteFile(path, embeddedDatabaseBytes, 0600)
}

func openDB(cfg config) (*sql.DB, error) {
	dbPath, err := filepath.Abs(cfg.db.path)
	if err != nil {
		return nil, err
	}

	if err := hydrateDB(dbPath); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

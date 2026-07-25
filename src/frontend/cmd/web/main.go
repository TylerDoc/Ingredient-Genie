package main

import (
	"flag"
	"html/template"
	"log/slog"
	"os"

	"github.com/michaelgov-ctrl/Ingredient-Genie-frontend/internal/data"

	"github.com/go-playground/form/v4"
)

type config struct {
	port        int
	logLevel    string
	mealsApiUri string
}

type application struct {
	config        config
	logger        *slog.Logger
	models        data.Models
	templateCache map[string]*template.Template
	formDecoder   *form.Decoder
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4243, "API server port")
	flag.StringVar(&cfg.logLevel, "log-level", "error", "Logging level (trace|debug|info|warning|error)")

	flag.StringVar(&cfg.mealsApiUri, "meals-api-uri", "http://localhost:4242", "URI of the meals API to target")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		config:        cfg,
		logger:        logger,
		models:        data.NewModels(logger, cfg.mealsApiUri),
		templateCache: templateCache,
		formDecoder:   form.NewDecoder(),
	}

	if err := app.serve(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

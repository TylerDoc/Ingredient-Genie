package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/michaelgov-ctrl/Ingredient-Genie-frontend/internal/data"
	"github.com/michaelgov-ctrl/Ingredient-Genie-frontend/ui"
)

type templateData struct {
	Form        any
	Meal        data.Meal
	Meals       []data.Meal
	MealMatches []data.MealMatch
	Metadata    data.Metadata
	SortTypes   []data.SortType
	CSRFToken   string
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func previousPage(page int) int {
	return page - 1
}

func nextPage(page int) int {
	return page + 1
}

var functions = template.FuncMap{
	"humanDate":    humanDate,
	"previousPage": previousPage,
	"nextPage":     nextPage,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.tmpl.html",
			"html/partials/*.html",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

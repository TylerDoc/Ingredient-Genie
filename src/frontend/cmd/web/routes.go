package main

import (
	"net/http"

	"github.com/michaelgov-ctrl/Ingredient-Genie-frontend/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	mux.HandleFunc("GET /{$}", app.home)

	mux.HandleFunc("GET /meal/create", app.mealCreate)
	mux.HandleFunc("POST /meal/create", app.mealCreatePost)

	mux.HandleFunc("GET /meal/view/{id}", app.mealView)
	mux.HandleFunc("POST /meal/view/{id}/update", app.mealUpdate)
	mux.HandleFunc("POST /meal/view/{id}/delete", app.mealDelete)

	mux.HandleFunc("GET /meal/search", app.searchMealsByIngredients)
	mux.HandleFunc("POST /meal/search", app.searchMealsByIngredientsPost)

	return app.recoverPanic(app.logRequest(commonHeaders(noSurf(mux))))
}

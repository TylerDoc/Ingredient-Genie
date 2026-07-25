package main

import (
	"net/http"
)

// # out of place testing note
//
//	$body = @{
//		"ingredients" = @(
//			"Garlic",
//			"Red Onions",
//			"Vegetable Oil",
//			"Lime"
//		)
//		"filters" = @{
//			"page" = 1
//			"pageSize" = 10
//			"sort" = "-ratio"
//		}
//	} | ConvertTo-Json
//
// $resp = irm http://localhost:4242/v1/meals/search -Method POST -Body $body -ContentType "application/json"
// # use iwr to check metadata
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)
	mux.HandleFunc("POST /v1/meals/create", app.createMealHandler)
	mux.HandleFunc("POST /v1/meals/get", app.getMealHandler)
	mux.HandleFunc("PUT /v1/meals/update", app.updateMealHandler)
	mux.HandleFunc("DELETE /v1/meals/delete", app.deleteMealHandler)
	mux.HandleFunc("GET /v1/meals/sort", app.mealSortTypesHandler)
	mux.HandleFunc("POST /v1/meals/list", app.listMealsHandler)
	mux.HandleFunc("POST /v1/meals/search", app.searchMealByIngredientsHandler)

	return app.recoverPanic(app.enableCORS(app.logRequest(app.rateLimit(mux))))
}

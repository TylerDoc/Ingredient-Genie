package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/michaelgov-ctrl/Ingredient-Genie-backend/internal/data"
	"github.com/michaelgov-ctrl/Ingredient-Genie-backend/internal/validator"
)

type envelope map[string]any

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.writeJSON(w, http.StatusOK, envelope{"status": "available"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createMealHandler(w http.ResponseWriter, r *http.Request) {
	var meal data.Meal

	if err := app.readJSON(w, r, &meal); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateMeal(v, meal); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := app.models.Meals.Create(ctx, meal)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/meals/%d", id))

	if err := app.writeJSON(w, http.StatusCreated, envelope{"id": id}, headers); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getMealHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID int64 `json:"id"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.ID > 0, "id", "must be greater than 0")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	meal, err := app.models.Meals.Get(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"meal": meal}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMealHandler(w http.ResponseWriter, r *http.Request) {
	var meal data.Meal

	if err := app.readJSON(w, r, &meal); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(meal.ID > 0, "id", "must be greater than 0")
	data.ValidateMeal(v, meal)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := app.models.Meals.Update(ctx, meal); err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err := app.writeJSON(w, http.StatusOK, envelope{"meal": meal}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMealHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID int64 `json:"id"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if v.Check(input.ID > 0, "id", "must be greater than 0"); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := app.models.Meals.Delete(ctx, input.ID); err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) mealSortTypesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: this is probably a good reason sort options should be enums..
	if err := app.writeJSON(w, http.StatusOK, envelope{"sortTypes": data.MealSortOptions}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listMealsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters `json:"filters"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	meals, metadata, err := app.models.Meals.GetAll(ctx, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"meals": meals, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) searchMealByIngredientsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Ingredients  []string `json:"ingredients"`
		data.Filters `json:"filters"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateIngredientSearch(v, input.Ingredients); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mealMatches, metadata, err := app.models.Meals.FindByIngredients(ctx, input.Ingredients, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// fmt.Println(metadata, len(meals), "\n", meals[len(meals)-1])
	if err := app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "mealMatches": mealMatches}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	var max_bytes int64 = 1_048_576 // 1mb
	r.Body = http.MaxBytesReader(w, r.Body, max_bytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var (
			syntaxError           *json.SyntaxError
			unmarshalTypeError    *json.UnmarshalTypeError
			invalidUnmarshalError *json.InvalidUnmarshalError
			maxBytesError         *http.MaxBytesError
		)

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON at type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldname := strings.TrimPrefix(err.Error(), "json: unknown")
			return fmt.Errorf("body contains unknown key %s", fieldname)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes, that's one chonky meme", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

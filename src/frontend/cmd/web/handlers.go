package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/michaelgov-ctrl/Ingredient-Genie-frontend/internal/data"
	"github.com/michaelgov-ctrl/Ingredient-Genie-frontend/internal/validator"
)

const (
	defaultMealSearchPageSize = 9
	defaultMealSearchSort     = "-ratio"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 18

	page, err := lazyDefault(r, "page", page)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	pageSize, err = lazyDefault(r, "pageSize", pageSize)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	resp, err := app.models.Meals.GetMealList(data.Filters{
		Page:     page,
		PageSize: pageSize,
		Sort:     "name",
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	templateData := app.newTemplateData(r)
	templateData.Meals = resp.Meals
	templateData.Metadata = resp.Metadata

	app.render(w, r, http.StatusOK, "home.tmpl.html", templateData)
}

type mealForm struct {
	ID            int64                 `form:"id"`
	Name          string                `form:"name"`
	AlternateName string                `form:"alternate_name"`
	Category      string                `form:"category"`
	Area          string                `form:"area"`
	Country       string                `form:"country"`
	Instructions  string                `form:"instructions"`
	YoutubeURL    string                `form:"youtube_url"`
	SourceURL     string                `form:"source_url"`
	Ingredients   []data.MealIngredient `form:"ingredients"`

	validator.Validator `form:"-"`
}

func mealToForm(meal data.Meal) mealForm {
	return mealForm{
		ID:            meal.ID,
		Name:          meal.Name,
		AlternateName: meal.AlternateName,
		Category:      meal.Category,
		Area:          meal.Area,
		Country:       meal.Country,
		Instructions:  meal.Instructions,
		YoutubeURL:    meal.YoutubeURL,
		SourceURL:     meal.SourceURL,
		Ingredients:   meal.Ingredients,
	}
}

func (f mealForm) Meal() data.Meal {
	ingredients := make([]data.MealIngredient, len(f.Ingredients))

	for i, ingredient := range f.Ingredients {
		ingredient.Position = int64(i + 1)
		ingredients[i] = ingredient
	}

	return data.Meal{
		ID:            f.ID,
		Name:          f.Name,
		AlternateName: f.AlternateName,
		Category:      f.Category,
		Area:          f.Area,
		Country:       f.Country,
		Instructions:  f.Instructions,
		YoutubeURL:    f.YoutubeURL,
		SourceURL:     f.SourceURL,
		Ingredients:   ingredients,
	}
}

func (app *application) mealCreate(w http.ResponseWriter, r *http.Request) {
	form := mealForm{}

	templateData := app.newTemplateData(r)
	templateData.Form = form

	app.render(w, r, http.StatusOK, "meal.tmpl.html", templateData)
}

func (app *application) mealCreatePost(w http.ResponseWriter, r *http.Request) {
	var form mealForm

	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	meal := form.Meal()

	if data.ValidateMeal(&form.Validator, meal); !form.Valid() {
		templateData := app.newTemplateData(r)
		templateData.Form = form

		app.render(w, r, http.StatusUnprocessableEntity, "meal.tmpl.html", templateData)
		return
	}

	id, err := app.models.Meals.CreateMeal(meal)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/meal/view/%d", id), http.StatusSeeOther)
}

func (app *application) mealView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	meal, err := app.models.Meals.GetMeal(id)
	if err != nil {
		if errors.Is(err, data.ErrNoMeal) {
			http.NotFound(w, r)
			return
		}

		app.serverError(w, r, err)
		return
	}

	templateData := app.newTemplateData(r)
	templateData.Form = mealToForm(meal)

	app.render(w, r, http.StatusOK, "meal.tmpl.html", templateData)
}

func (app *application) mealUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	var form mealForm

	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.ID = int64(id)

	meal := form.Meal()
	if data.ValidateMeal(&form.Validator, meal); !form.Valid() {
		templateData := app.newTemplateData(r)
		templateData.Form = form

		app.render(w, r, http.StatusUnprocessableEntity, "meal.tmpl.html", templateData)
		return
	}

	if err := app.models.Meals.UpdateMeal(meal); err != nil {
		if errors.Is(err, data.ErrNoMeal) {
			http.NotFound(w, r)
			return
		}

		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/meal/view/%d", id), http.StatusSeeOther)
}

func (app *application) mealDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	if err := app.models.Meals.DeleteMeal(id); err != nil {
		if errors.Is(err, data.ErrNoMeal) {
			http.NotFound(w, r)
			return
		}

		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type mealSearchForm struct {
	Ingredients         []string `form:"ingredients"`
	Page                int      `form:"page"`
	PageSize            int      `form:"pagesize"`
	Sort                string   `form:"sort"`
	validator.Validator `form:"-"`
}

func (app *application) searchMealsByIngredients(w http.ResponseWriter, r *http.Request) {
	sortTypes, err := app.models.Meals.GetSortTypes()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// defaults
	form := mealSearchForm{
		Page:     1,
		PageSize: defaultMealSearchPageSize,
		Sort:     defaultMealSearchSort,
	}

	query := r.URL.Query()

	form.Ingredients = normalizeIngredients(query["ingredients"])

	if value := query.Get("page"); value != "" {
		form.Page, err = strconv.Atoi(value)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	if value := query.Get("pagesize"); value != "" {
		form.PageSize, err = strconv.Atoi(value)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	if value := query.Get("sort"); value != "" {
		form.Sort = value
	}

	templateData := app.newTemplateData(r)
	templateData.Form = form
	templateData.SortTypes = sortTypes

	// first visit base case.
	if len(form.Ingredients) == 0 {
		app.render(w, r, http.StatusOK, "ingredientsearch.tmpl.html", templateData)
		return
	}

	form.CheckField(form.Page >= 1, "page", "Page must be greater than zero")
	form.CheckField(form.PageSize >= 1 && form.PageSize <= 100, "pagesize", "Page size must be between 1 and 100")

	if !form.Valid() {
		templateData.Form = form

		app.render(w, r, http.StatusUnprocessableEntity, "ingredientsearch.tmpl.html", templateData)
		return
	}

	req := data.IngredientMealSearchRequest{
		Ingredients: form.Ingredients,
		Filters: data.Filters{
			Page:     form.Page,
			PageSize: form.PageSize,
			Sort:     data.SortType(form.Sort),
		},
	}

	resp, err := app.models.Meals.SearchByIngredients(req)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	templateData.MealMatches = resp.MealMatches
	templateData.Metadata = resp.Metadata

	app.render(w, r, http.StatusOK, "ingredientsearch.tmpl.html", templateData)
}

func (app *application) searchMealsByIngredientsPost(w http.ResponseWriter, r *http.Request) {
	var form mealSearchForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.Ingredients = normalizeIngredients(form.Ingredients)

	if form.PageSize == 0 {
		form.PageSize = defaultMealSearchPageSize
	}

	if form.Sort == "" {
		form.Sort = defaultMealSearchSort
	}

	// A new search always starts on page 1.
	form.Page = 1

	form.CheckField(validator.NotEmpty(form.Ingredients), "ingredients", "At least one ingredient must be provided")
	form.CheckField(form.PageSize >= 1 && form.PageSize <= 100, "pagesize", "Page size must be between 1 and 100")

	if !form.Valid() {
		sortTypes, err := app.models.Meals.GetSortTypes()
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		templateData := app.newTemplateData(r)
		templateData.Form = form
		templateData.SortTypes = sortTypes

		app.render(w, r, http.StatusUnprocessableEntity, "ingredientsearch.tmpl.html", templateData)

		return
	}

	query := url.Values{}

	for _, ingredient := range form.Ingredients {
		query.Add("ingredients", ingredient)
	}

	query.Set("page", strconv.Itoa(form.Page))
	query.Set("pagesize", strconv.Itoa(form.PageSize))
	query.Set("sort", form.Sort)

	http.Redirect(w, r, "/meal/search?"+query.Encode(), http.StatusSeeOther)
}

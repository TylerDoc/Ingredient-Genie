package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/michaelgov-ctrl/Ingredient-Genie-frontend/internal/validator"
)

// TODO: some kind of validation on addr
type MealsClient struct {
	logger                 *slog.Logger
	addr                   string
	healthcheckEndpoint    string
	mealsCreateEndpoint    string
	mealsGetEndpoint       string
	mealsUpdateEndpoint    string
	mealsDeleteEndpoint    string
	mealsListEndpoint      string
	mealsSearchEndpoint    string
	mealsSortTypesEndpoint string
	httpClient             *http.Client
}

func NewMealsClient(logger *slog.Logger, addr string) MealsClient {
	version := "/v1"

	client := MealsClient{
		logger:                 logger,
		addr:                   addr,
		healthcheckEndpoint:    version + "/healthcheck",
		mealsCreateEndpoint:    version + "/meals/create",
		mealsGetEndpoint:       version + "/meals/get",
		mealsUpdateEndpoint:    version + "/meals/update",
		mealsDeleteEndpoint:    version + "/meals/delete",
		mealsListEndpoint:      version + "/meals/list",
		mealsSearchEndpoint:    version + "/meals/search",
		mealsSortTypesEndpoint: version + "/meals/sort",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	go func() {
		// TODO: this could bubble up to the website or something as well as logging.
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := client.CheckHealth(); err != nil {
				client.logger.Error("meals API health check failed", "error", err)
			}
		}
	}()

	return client
}

func (mc MealsClient) CheckHealth() error {
	req, err := http.NewRequest(http.MethodGet, mc.addr+mc.healthcheckEndpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")

	resp, err := mc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("meals API health check returned status %d", resp.StatusCode)
	}

	var Response struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&Response); err != nil {
		return err
	}

	if Response.Status != "available" {
		return fmt.Errorf("meals api failed health check with: %s", Response.Status)
	}

	return nil
}

func ValidateMeal(v *validator.Validator, meal Meal) {
	v.CheckField(strings.TrimSpace(meal.Name) != "", "name", "must be provided")
	v.CheckField(len(meal.Name) <= 200, "name", "must not be more than 200 characters")

	v.CheckField(len(meal.AlternateName) <= 200, "alternateName", "must not be more than 200 characters")
	v.CheckField(len(meal.Category) <= 100, "category", "must not be more than 100 characters")
	v.CheckField(len(meal.Area) <= 100, "area", "must not be more than 100 characters")
	v.CheckField(len(meal.Country) <= 100, "country", "must not be more than 100 characters")

	v.CheckField(strings.TrimSpace(meal.Instructions) != "", "instructions", "must be provided")

	v.CheckField(len(meal.Ingredients) > 0, "ingredients", "must contain at least one ingredient")

	positions := make(map[int64]struct{})

	for _, ingredient := range meal.Ingredients {
		v.CheckField(strings.TrimSpace(ingredient.Name) != "", "ingredients", "ingredient names must be provided")
		v.CheckField(len(ingredient.Name) <= 200, "ingredients", "ingredient names must not be more than 200 characters")

		v.CheckField(ingredient.Position > 0, "ingredients", "must contain valid ingredient positions")
		if _, exists := positions[ingredient.Position]; exists {
			v.AddFieldError("ingredients", "must not contain duplicate ingredient positions")
		}
		positions[ingredient.Position] = struct{}{}
	}
}

func (mc MealsClient) CreateMeal(meal Meal) (int, error) {
	body, err := json.Marshal(meal)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, mc.addr+mc.mealsCreateEndpoint, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := mc.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("create meal: unexpected status code %d", resp.StatusCode)
	}

	var response struct {
		ID int `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, err
	}

	return response.ID, nil
}

func (mc MealsClient) GetMeal(id int) (Meal, error) {
	input := struct {
		ID int `json:"id"`
	}{
		ID: id,
	}

	body, err := json.Marshal(input)
	if err != nil {
		return Meal{}, err
	}

	req, err := http.NewRequest(http.MethodPost, mc.addr+mc.mealsGetEndpoint, bytes.NewReader(body))
	if err != nil {
		return Meal{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := mc.httpClient.Do(req)
	if err != nil {
		return Meal{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Meal{}, ErrNoMeal
	}

	if resp.StatusCode != http.StatusOK {
		return Meal{}, fmt.Errorf("failed to get meal with http status code: %d", resp.StatusCode)
	}

	var response struct {
		Meal Meal `json:"meal"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Meal{}, err
	}

	return response.Meal, nil
}

func (mc MealsClient) UpdateMeal(meal Meal) error {
	body, err := json.Marshal(meal)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, mc.addr+mc.mealsUpdateEndpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := mc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNoMeal
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update meal: unexpected status code %d", resp.StatusCode)
	}

	return nil
}

func (mc MealsClient) DeleteMeal(id int) error {
	input := struct {
		ID int `json:"id"`
	}{
		ID: id,
	}

	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, mc.addr+mc.mealsDeleteEndpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := mc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNoMeal
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete meal: unexpected status code %d", resp.StatusCode)
	}

	return nil
}

func (mc MealsClient) GetMealList(filters Filters) (MealListResponse, error) {
	input := struct {
		Filters Filters `json:"filters"`
	}{
		Filters: filters,
	}

	body, err := json.Marshal(input)
	if err != nil {
		return MealListResponse{}, err
	}

	req, err := http.NewRequest(http.MethodPost, mc.addr+mc.mealsListEndpoint, bytes.NewReader(body))
	if err != nil {
		return MealListResponse{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := mc.httpClient.Do(req)
	if err != nil {
		return MealListResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return MealListResponse{}, fmt.Errorf("failed to get meal list with http status code: %d", resp.StatusCode)
	}

	var response MealListResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return MealListResponse{}, err
	}

	return response, nil
}

func (mc MealsClient) GetSortTypes() ([]SortType, error) {
	req, err := http.NewRequest(http.MethodGet, mc.addr+mc.mealsSortTypesEndpoint, nil)
	if err != nil {
		return []SortType{}, err
	}

	req.Header.Add("Accept", "application/json")

	resp, err := mc.httpClient.Do(req)
	if err != nil {
		return []SortType{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []SortType{}, fmt.Errorf("failed to enumerate sort types with http status code: %d", resp.StatusCode)
	}

	var Response struct {
		SortTypes []SortType `json:"sortTypes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&Response); err != nil {
		return []SortType{}, err
	}

	return Response.SortTypes, nil
}

func (mc MealsClient) SearchByIngredients(body IngredientMealSearchRequest) (MealSearchResponse, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return MealSearchResponse{}, err
	}

	req, err := http.NewRequest(http.MethodPost, mc.addr+mc.mealsSearchEndpoint, bytes.NewReader(payload))
	if err != nil {
		return MealSearchResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := mc.httpClient.Do(req)
	if err != nil {
		return MealSearchResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return MealSearchResponse{}, fmt.Errorf("failed to search ingredients with http status code: %d", resp.StatusCode)
	}

	var msr MealSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&msr); err != nil {
		return MealSearchResponse{}, err

	}

	return msr, nil
}

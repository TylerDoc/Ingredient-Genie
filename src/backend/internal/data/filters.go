package data

import (
	"github.com/michaelgov-ctrl/Ingredient-Genie-backend/internal/validator"
)

type Filters struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Sort     string `json:"sort"`
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than 0")
	v.Check(f.Page <= 1_000, "page", "must be less than or equal to 1,000") // this may cause an unexpected issue if we surpass 1000 pages, unlikely..
	v.Check(f.PageSize > 0, "pageSize", "must be greater than 0")
	v.Check(f.PageSize <= 50, "pageSize", "must be less than or equal to 50")

	_, validSort := MealSortStmts[f.Sort]
	v.Check(validSort, "sort", "invalid sort value")
}

// TODO: should these be enums?? but it generates a lot of boilerplate..
var MealSortStmts = map[string]string{
	"id":     "MealId ASC",
	"-id":    "MealId DESC",
	"ratio":  "MatchRatio ASC, MealId ASC",
	"-ratio": "MatchRatio DESC, MealId ASC",
	"name":   "Name COLLATE NOCASE ASC, MealId ASC",
	"-name":  "Name COLLATE NOCASE DESC, MealId ASC",
}

func (f Filters) orderBy() string {
	expression, ok := MealSortStmts[f.Sort]
	if !ok {
		panic("unsafe sort parameter: " + f.Sort)
	}

	return expression
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

type Metadata struct {
	CurrentPage  int `json:"currentPage"`
	PageSize     int `json:"pageSize"`
	FirstPage    int `json:"firstPage"`
	LastPage     int `json:"lastPage"`
	TotalRecords int `json:"totalRecords"`
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (totalRecords + pageSize - 1) / pageSize,
		TotalRecords: totalRecords,
	}
}

package data

import (
	"testing"

	"github.com/michaelgov-ctrl/Ingredient-Genie-backend/internal/validator"
)

// test 08
// Unsupported sort option
func TestValidateFilters_UnsupportedSort(t *testing.T) {
	v := validator.New()

	filters := Filters{
		Page:     1,
		PageSize: 10,
		Sort:     "sideways",
	}

	ValidateFilters(v, filters)

	if v.Valid() {
		t.Fatal("expected unsupported sort option to fail validation")
	}

	if _, ok := v.Errors["sort"]; !ok {
		t.Error("expected validation error for sort")
	}
}

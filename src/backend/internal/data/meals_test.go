package data

import (
	"context"
	"database/sql"
	"math"
	"testing"

	"github.com/michaelgov-ctrl/Ingredient-Genie-backend/internal/validator"
	_ "modernc.org/sqlite"
)

func newTestMealModel(t *testing.T) MealModel {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	db.SetMaxOpenConns(1)

	t.Cleanup(func() {
		db.Close()
	})

	statements := []string{
		`
		CREATE TABLE Meal (
			MealId INTEGER PRIMARY KEY,
			Name TEXT,
			AlternateName TEXT,
			Category TEXT,
			Area TEXT,
			Country TEXT,
			Instructions TEXT,
			YoutubeUrl TEXT,
			SourceUrl TEXT
		);
		`,
		`
		CREATE TABLE Ingredient (
			IngredientId INTEGER PRIMARY KEY,
			Name TEXT,
			NormalizedName TEXT UNIQUE
		);
		`,
		`
		CREATE TABLE MealIngredient (
			MealId INTEGER NOT NULL,
			IngredientId INTEGER NOT NULL,
			Position INTEGER NOT NULL,
			MeasureText TEXT,
			PRIMARY KEY (MealId, Position)
		);
		`,
		`
		INSERT INTO Meal (
			MealId,
			Name
		)
		VALUES
			(1, 'Chicken Garlic Bowl'),
			(2, 'Garlic Rice'),
			(3, 'Beef Tacos');
		`,
		`
		INSERT INTO Ingredient (
			IngredientId,
			Name,
			NormalizedName
		)
		VALUES
			(1, 'Chicken', 'chicken'),
			(2, 'Garlic', 'garlic'),
			(3, 'Rice', 'rice'),
			(4, 'Beef', 'beef'),
			(5, 'Tortilla', 'tortilla'),
			(6, 'Onion', 'onion');
		`,
		`
		INSERT INTO MealIngredient (
			MealId,
			IngredientId,
			Position,
			MeasureText
		)
		VALUES
			(1, 1, 1, '1 lb'),
			(1, 2, 2, '2 cloves'),
			(1, 3, 3, '1 cup'),

			(2, 2, 1, '2 cloves'),
			(2, 3, 2, '1 cup'),

			(3, 4, 1, '1 lb'),
			(3, 5, 2, '3'),
			(3, 6, 3, '1');
		`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("setting up test database: %v", err)
		}
	}

	return MealModel{DB: db}
}

func testFilters(sort string) Filters {
	return Filters{
		Page:     1,
		PageSize: 10,
		Sort:     sort,
	}
}

func findMealMatch(t *testing.T, matches []MealMatch, name string) MealMatch {
	t.Helper()

	for _, match := range matches {
		if match.Meal.Name == name {
			return match
		}
	}

	t.Fatalf("meal %q not found", name)

	return MealMatch{}
}

// test 01
// Multiple ingredient search
func TestFindByIngredients_MultipleIngredients(t *testing.T) {
	model := newTestMealModel(t)

	matches, _, err := model.FindByIngredients(context.Background(), []string{"chicken", "garlic"}, testFilters("-ratio"))
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 2 {
		t.Fatalf("expected 2 matching meals, got %d", len(matches))
	}

	chickenBowl := findMealMatch(t, matches, "Chicken Garlic Bowl")

	if chickenBowl.MatchedIngredientCount != 2 {
		t.Errorf("expected 2 matched ingredients, got %d", chickenBowl.MatchedIngredientCount)
	}

	if chickenBowl.TotalIngredientCount != 3 {
		t.Errorf("expected 3 total ingredients, got %d", chickenBowl.TotalIngredientCount)
	}
}

// test 02
// Unknown ingredient
func TestFindByIngredients_UnknownIngredient(t *testing.T) {
	model := newTestMealModel(t)

	matches, metadata, err := model.FindByIngredients(context.Background(), []string{"papaya"}, testFilters("-ratio"))
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 0 {
		t.Errorf("expected no matching meals, got %d", len(matches))
	}

	if metadata.TotalRecords != 0 {
		t.Errorf("expected 0 total records, got %d", metadata.TotalRecords)
	}
}

// test 03
// No ingredients
func TestValidateIngredientSearch_NoIngredients(t *testing.T) {
	v := validator.New()

	ValidateIngredientSearch(v, []string{})

	if v.Valid() {
		t.Fatal("expected empty ingredient search to fail validation")
	}

	if _, ok := v.Errors["ingredients"]; !ok {
		t.Error("expected validation error for ingredients")
	}
}

// test 04
// Duplicate ingredients
func TestFindByIngredients_DuplicateIngredients(t *testing.T) {
	model := newTestMealModel(t)

	normalMatches, _, err := model.FindByIngredients(context.Background(), []string{"chicken", "garlic"}, testFilters("-ratio"))
	if err != nil {
		t.Fatal(err)
	}

	duplicateMatches, _, err := model.FindByIngredients(context.Background(), []string{"chicken", "chicken", "garlic"}, testFilters("-ratio"))
	if err != nil {
		t.Fatal(err)
	}

	normal := findMealMatch(t, normalMatches, "Chicken Garlic Bowl")
	duplicate := findMealMatch(t, duplicateMatches, "Chicken Garlic Bowl")

	if duplicate.MatchedIngredientCount != normal.MatchedIngredientCount {
		t.Errorf("duplicate ingredient changed matched count: expected %d, got %d", normal.MatchedIngredientCount, duplicate.MatchedIngredientCount)
	}

	if duplicate.MatchRatio != normal.MatchRatio {
		t.Errorf("duplicate ingredient changed match ratio: expected %f, got %f", normal.MatchRatio, duplicate.MatchRatio)
	}
}

// test 05
// Partial ingredient match
func TestFindByIngredients_PartialMatch(t *testing.T) {
	model := newTestMealModel(t)

	matches, _, err := model.FindByIngredients(context.Background(), []string{"chicken"}, testFilters("-ratio"))
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 matching meal, got %d", len(matches))
	}

	match := matches[0]

	if match.Meal.Name != "Chicken Garlic Bowl" {
		t.Errorf("expected Chicken Garlic Bowl, got %q", match.Meal.Name)
	}

	expectedRatio := 1.0 / 3.0

	if math.Abs(match.MatchRatio-expectedRatio) > 0.000001 {
		t.Errorf("expected match ratio %f, got %f", expectedRatio, match.MatchRatio)
	}

	if match.MatchedIngredientCount != 1 {
		t.Errorf("expected 1 matched ingredient, got %d", match.MatchedIngredientCount)
	}

	if match.TotalIngredientCount != 3 {
		t.Errorf("expected 3 total ingredients, got %d", match.TotalIngredientCount)
	}
}

// test 06
// Complete match percentage
func TestFindByIngredients_CompleteMatch(t *testing.T) {
	model := newTestMealModel(t)

	matches, _, err := model.FindByIngredients(context.Background(), []string{"garlic", "rice"}, testFilters("-ratio"))
	if err != nil {
		t.Fatal(err)
	}

	garlicRice := findMealMatch(t, matches, "Garlic Rice")
	if garlicRice.MatchRatio != 1.0 {
		t.Errorf("expected complete match ratio of 1.0, got %f", garlicRice.MatchRatio)
	}

	if garlicRice.MatchedIngredientCount != garlicRice.TotalIngredientCount {
		t.Errorf("expected matched count to equal total count: %d != %d", garlicRice.MatchedIngredientCount, garlicRice.TotalIngredientCount)
	}

	for _, match := range matches {
		if match.MatchRatio > 1.0 {
			t.Errorf("meal %q has match ratio greater than 1.0: %f", match.Meal.Name, match.MatchRatio)
		}
	}
}

// test 07
// Sort by match ratio
func TestFindByIngredients_SortByMatchRatioDescending(t *testing.T) {
	model := newTestMealModel(t)

	matches, _, err := model.FindByIngredients(context.Background(), []string{"garlic", "rice"}, testFilters("-ratio"))
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 2 {
		t.Fatalf("expected 2 matching meals, got %d", len(matches))
	}

	for i := 1; i < len(matches); i++ {
		if matches[i-1].MatchRatio < matches[i].MatchRatio {
			t.Errorf("results are not sorted descending: %f came before %f", matches[i-1].MatchRatio, matches[i].MatchRatio)
		}
	}

	if matches[0].Meal.Name != "Garlic Rice" {
		t.Errorf("expected highest match to be Garlic Rice, got %q", matches[0].Meal.Name)
	}

	if matches[0].MatchRatio != 1.0 {
		t.Errorf("expected highest match ratio to be 1.0, got %f", matches[0].MatchRatio)
	}
}

# Ingredient Genie

Ingredient Genie is a web application for finding meals based on ingredients users already have available.

The application helps reduce food waste, avoid unnecessary grocery purchases, and simplify meal planning.

## Overview

Ingredient Genie uses a client-server architecture with two standalone Go applications:

* Go web frontend
* Go JSON REST API backend
* SQLite database
* HTML templates, CSS, and JavaScript

The frontend communicates with the backend through HTTP and JSON. The backend API can also be accessed independently.

```text
Browser <-> Go Frontend <-> JSON REST API <-> Go Backend <-> SQLite
```

The project contains separate Go modules for the frontend and backend:

```text
src/
├── backend/
└── frontend/
```

## Features

Ingredient Genie supports:

* Searching meals using a variable-length ingredient list
* Ranking results based on matching ingredients
* Displaying matching and missing ingredients
* Sorting and pagination
* Creating meals
* Viewing meals
* Updating meals
* Deleting meals
* Dynamic ingredient fields using JavaScript
* Input validation and error handling
* Backend health checks
* Responsive browser UI

## Data

Meal information is stored in SQLite using three primary tables:

* `Meal`
* `Ingredient`
* `MealIngredient`

Ingredients are stored separately and shared between meals. `MealIngredient` stores the relationship between a meal and its ingredients, including measurement and ingredient position.

## Backend API

The backend provides JSON endpoints for:

* Health checks
* Meal creation
* Meal retrieval
* Meal updates
* Meal deletion
* Meal listing
* Ingredient-based searches
* Available search sort options

Example search request:

```json
{
  "ingredients": [
    "garlic",
    "chicken",
    "rice"
  ],
  "page": 1,
  "pageSize": 10,
  "sort": "-ratio"
}
```

The API validates requests and returns appropriate JSON responses and HTTP status codes.

## Frontend

The frontend uses Go HTML templates for server-side rendering.

JavaScript is used for lightweight client-side behavior such as:

* Adding ingredient fields
* Removing ingredient fields
* Reindexing ingredient fields
* Confirming meal deletion

The same meal form is used for both creating new meals and viewing or updating existing meals.

## Running the Project

Run the backend:

```bash
cd src/backend
go run ./cmd/api
```

Run the frontend:

```bash
cd src/frontend
go run ./cmd/web
```

The frontend must be configured with the address of the backend API.

## Development Workflow

Development uses Git and GitHub feature branches.

```bash
git clone git@github.com:TylerDoc/Ingredient-Genie.git
cd Ingredient-Genie

git checkout -b my-feature-branch

git add .
git commit -m "Add feature"
git push -u origin my-feature-branch
```

Changes are merged into `main` through GitHub pull requests.

To update a feature branch:

```bash
git checkout main
git pull origin main

git checkout my-feature-branch
git rebase main
```

## CI/CD and Releases

GitHub Actions and GoReleaser are used to build and publish releases.

GoReleaser builds both Go applications:

* `ingredient-genie-frontend`
* `ingredient-genie-backend`

Release binaries are built for:

* Linux - AMD64 & ARM64
* Windows - AMD64 & ARM64
* macOS - AMD64 & ARM64

The release workflow is:

```text
Git Tag -> GitHub Actions -> GoReleaser -> GitHub Release
```

This provides repeatable builds of both applications and publishes the compiled binaries through GitHub Releases.

## Go References

The project was written with an emphasis on conventional and idiomatic Go.

Alex Edwards' books were used as development references:

* *Let's Go*
* *Let's Go Further*

They were used for guidance on topics including routing, middleware, templates, validation, database access, HTTP clients, JSON APIs, and general Go web application structure.

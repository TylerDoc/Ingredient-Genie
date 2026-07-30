# Ingredient Genie

Visit [Our Site](https://ingredient-genie.com)

Ingredient Genie is a web application for finding meals based on ingredients users already have available.

The application helps reduce food waste, avoid unnecessary grocery purchases, and simplify meal planning.

[TheMealDB](https://www.themealdb.com/) was used to gather seed data for the database.

## Host your own Ingredient-Genie Instance

* Navigate to the [releases page](https://github.com/TylerDoc/Ingredient-Genie/releases)
* Scroll to the bottom and download the backend and frontend archives for your machines architecture. (you may have to expand the asset list)
* Extract the archives.
* Run the extracted backend binary.
* Run the extracted frontend binary.
* Navigate to localhost:4243 by default to browse meals.
* For further options read the included readme's or run the binaries with the `-help` flag.

## Docs directory

The docs directory contains further documentation for our deployment and development processes.

```text
docs/
├── deployment.md
└── development.md
```

## Go References

The project was written with an emphasis on conventional and idiomatic Go.

Alex Edwards' books were used as development references:

* *Let's Go*
* *Let's Go Further*

They were used for guidance on topics including routing, middleware, templates, validation, database access, HTTP clients, JSON APIs, and general Go web application structure.

## AI Disclosure

AI was specifically used to:
* Chunks of the code used to seed the database in `Generate-Database.ps1`
* Generate the CSS for the front end
* The regex for the release matching in the Ansible playbook
* The algorithm for matching a list of ingredients with meals
* Parts of this readme
* and to answer questions to guide further learning
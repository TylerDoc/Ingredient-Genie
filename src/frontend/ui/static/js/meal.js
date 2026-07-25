document.addEventListener("DOMContentLoaded", () => {
    const container = document.getElementById("meal-ingredient-inputs");
    const addButton = document.getElementById("add-meal-ingredient");

    if (container && addButton) {
        function reindexIngredients() {
            const rows = container.querySelectorAll(".meal-ingredient-row");

            rows.forEach((row, index) => {
                const ingredientName = row.querySelector(
                    'input[name$=".Name"]'
                );

                const measureText = row.querySelector(
                    'input[name$=".MeasureText"]'
                );

                if (ingredientName) {
                    ingredientName.name = `ingredients[${index}].Name`;
                }

                if (measureText) {
                    measureText.name = `ingredients[${index}].MeasureText`;
                }
            });
        }

        addButton.addEventListener("click", () => {
            const row = document.createElement("div");
            row.classList.add("meal-ingredient-row");

            row.innerHTML = `
                <div>
                    <label>Ingredient</label>
                    <input
                        type="text"
                        placeholder="e.g. Garlic"
                        required>
                </div>

                <div>
                    <label>Measure</label>
                    <input
                        type="text"
                        placeholder="e.g. 2 cloves">
                </div>

                <button
                    type="button"
                    class="remove-meal-ingredient">
                    Remove
                </button>
            `;

            container.appendChild(row);
            reindexIngredients();
        });

        container.addEventListener("click", (event) => {
            const removeButton = event.target.closest(
                ".remove-meal-ingredient"
            );

            if (!removeButton) {
                return;
            }

            const row = removeButton.closest(".meal-ingredient-row");

            if (row) {
                row.remove();
                reindexIngredients();
            }
        });

        reindexIngredients();
    }

    const deleteForm = document.querySelector(".meal-delete-form");

    if (deleteForm) {
        deleteForm.addEventListener("submit", (event) => {
            const confirmed = confirm(
                "Are you sure you want to delete this meal?"
            );

            if (!confirmed) {
                event.preventDefault();
            }
        });
    }
});
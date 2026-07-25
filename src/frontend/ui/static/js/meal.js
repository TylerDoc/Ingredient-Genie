document.addEventListener("DOMContentLoaded", () => {
    const container = document.getElementById("create-ingredient-inputs");
    const addButton = document.getElementById("add-create-ingredient");

    if (!container || !addButton) {
        return;
    }

    function reindexIngredients() {
        const rows = container.querySelectorAll(".create-ingredient-row");

        rows.forEach((row, index) => {
            const ingredientName = row.querySelector(
                'input[name$=".IngredientName"]'
            );

            const measureText = row.querySelector(
                'input[name$=".MeasureText"]'
            );

            ingredientName.name =
                `ingredients[${index}].IngredientName`;

            measureText.name =
                `ingredients[${index}].MeasureText`;
        });
    }

    addButton.addEventListener("click", () => {
        const row = document.createElement("div");
        row.classList.add("create-ingredient-row");

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
                class="remove-create-ingredient">
                Remove
            </button>
        `;

        container.appendChild(row);
        reindexIngredients();
    });

    container.addEventListener("click", (event) => {
        const removeButton = event.target.closest(
            ".remove-create-ingredient"
        );

        if (!removeButton) {
            return;
        }

        const row = removeButton.closest(".create-ingredient-row");

        if (row) {
            row.remove();
            reindexIngredients();
        }
    });

    reindexIngredients();
});
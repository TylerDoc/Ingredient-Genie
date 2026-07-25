const ingredientInputs = document.getElementById("ingredient-inputs");
const addIngredientButton = document.getElementById("add-ingredient");

if (ingredientInputs && addIngredientButton) {
    addIngredientButton.addEventListener("click", () => {
        const index = ingredientInputs.children.length;

        const container = document.createElement("div");
        container.className = "ingredient-input";

        const label = document.createElement("label");
        label.htmlFor = `ingredient-${index}`;

        const input = document.createElement("input");
        input.type = "text";
        input.id = `ingredient-${index}`;
        input.name = "ingredients";
        input.placeholder = "e.g. Garlic";

        container.appendChild(label);
        container.appendChild(input);

        ingredientInputs.appendChild(container);
    });
}
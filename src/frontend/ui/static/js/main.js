// Footer
const footer = document.querySelector("footer");

if (footer) {
    const dayOfWeek = new Date().toLocaleString("en-US", {
        weekday: "long"
    });

    footer.textContent =
        dayOfWeek === "Thursday"
            ? "I never could get the hang of Thursdays"
            : `Have a nice ${dayOfWeek}`;
}

// Navigation
const navLinks = document.querySelectorAll("nav a");

for (const link of navLinks) {
    if (link.getAttribute("href") === window.location.pathname) {
        link.classList.add("live");
        break;
    }
}
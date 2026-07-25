// Footer
const footer = document.querySelector("footer");

if (footer) {
    const dayOfWeek = new Date().toLocaleString("en-US", {
        weekday: "long"
    });

    const quotes = {
        Monday: "Does someone have a case of the Mondays? ~ Office Space",
        Tuesday: "But for me, it was Tuesday. ~ Street Fighter",
        Wednesday: "On Wednesdays we wear pink. ~ Mean Girls",
        Thursday: "I never could get the hang of Thursdays. ~ Hitchhiker's Guide",
        Friday: "Thank God it's Friday ... again. Tomorrow is a rest day. ~ Farscape",
        Saturday: "A pretty nice little Saturday, we're going to go to Home Depot. ~ Old School",
        Sunday: "I once spent a year in Philadelphia, I think it was on a Sunday. ~ W. C. Fields"
    };

    footer.textContent = quotes[dayOfWeek] ?? "How did we get here?";
}

// Navigation
const navLinks = document.querySelectorAll("nav a");

for (const link of navLinks) {
    if (link.getAttribute("href") === window.location.pathname) {
        link.classList.add("live");
        break;
    }
}
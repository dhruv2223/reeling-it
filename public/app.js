import { HomePage } from "./components/HomePage.js";
import { MovieDetailsPage } from "./components/MovieDetailsPage.js";
import "./components/YoutubeEmbed.js";
import "./components/AnimatedLoading.js";
import { Router } from "./services/Router.js";
window.addEventListener("DOMContentLoaded", (event) => {
  app.Router.init();
});
window.app = {
  Router,
  showError: function (
    message = "There was an Error loading the page",
    goToHome = true,
  ) {
    document.getElementById("alert-mode").showModal();
    document.querySelector("#alert-mode p").textContent = message;
    if (goToHome) app.Router.go("/", false);
  },
  closeError: function () {
    document.getElementById("alert-mode").close();
  },
  search: function (event) {
    event.preventDefault();
    const q = document.querySelector("input[type=search]").value;
    console.log(q);
    app.Router.go("/movies?q=" + q);
  },

  searchOrderChange: (order) => {
    const urlParams = new URLSearchParams(window.location.search);
    const q = urlParams.get("q");
    const genre = urlParams.get("genre") ?? "";
    app.Router.go(`/movies?q=${q}&order=${order}&genre=${genre}`);
  },
  searchFilterChange: (genre) => {
    const urlParams = new URLSearchParams(window.location.search);
    const q = urlParams.get("q");
    const order = urlParams.get("order") ?? "";
    app.Router.go(`/movies?q=${q}&order=${order}&genre=${genre}`);
  },
};

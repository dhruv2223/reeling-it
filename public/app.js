import { HomePage } from "./components/HomePage.js";
import { MovieDetailsPage } from "./components/MovieDetailsPage.js";
import { API } from "./services/API.js";
import "./components/YoutubeEmbed.js";
import "./components/AnimatedLoading.js";
import { Router } from "./services/Router.js";
import Store from "./services/Store.js";
window.addEventListener("DOMContentLoaded", (event) => {
  app.Router.init();
});
window.app = {
  Router,
  Store,
  showError: function (
    message = "There was an Error loading the page",
    goToHome = false,
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
  register: async (event) => {
    event.preventDefault();
    const name = document.querySelector("#register-name").value;
    const email = document.querySelector("#register-email").value;
    const password = document.querySelector("#register-password").value;
    const confirmPassword = document.querySelector(
      "#register-password-confirmation",
    ).value;
    const errors = [];
    if (name.length < 2)
      errors.push("Name must be at least 2 characters long.");
    if (email.length < 5 || !email.includes("@")) {
      errors.push("Email must be a valid email address.");
    }
    if (password.length < 6) {
      errors.push("Password must be at least 6 characters long.");
    }
    if (password !== confirmPassword) {
      errors.push("Passwords do not match.");
    }
    if (errors.length > 0) {
      app.showError(errors.join(" "));
      return;
    } else {
      const response = await API.register(name, email, password);
      if (response.success) {
        app.Store.jwt = response.jwt;
        app.Router.go("/account/");
      } else {
        app.showError(response.message);
      }
    }
  },

  saveToCollection: async (movie_id, collection) => {
    if (app.Store.loggedIn) {
      try {
        const response = await API.saveToCollection(movie_id, collection);
        if (response.success) {
          switch (collection) {
            case "favorite":
              app.Router.go("/account/favorites");
              break;
            case "watchlist":
              app.Router.go("/account/watchlist");
          }
        } else {
          app.showError("We couldn't save the movie.");
        }
      } catch (e) {
        console.log(e);
      }
    } else {
      app.Router.go("/account/");
    }
  },
  login: async (event) => {
    event.preventDefault();
    const email = document.querySelector("#login-email").value;
    const password = document.querySelector("#login-password").value;
    const errors = [];
    if (email.length < 5 || !email.includes("@")) {
      errors.push("Email must be a valid email address.");
    }
    if (password.length < 6) {
      errors.push("Password must be at least 6 characters long.");
    }
    if (errors.length > 0) {
      app.showError(errors.join(" "));
      return;
    } else {
      const response = await API.login(email, password);
      if (response.success) {
        app.Store.jwt = response.jwt;
        app.Router.go("/account/");
      } else {
        app.showError(response.message);
      }
    }
  },
  logout: () => {
    Store.jwt = null;
    app.Router.go("/");
  },
};

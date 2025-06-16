export const API = {
  baseURL: "/api",
  getTopMovies: async function () {
    return await API.fetch("movies/top");
  },

  getRandomMovies: async function () {
    return await API.fetch("movies/random");
  },

  getGenres: async function () {
    return await API.fetch("/genres");
  },
  getMovieById: async function (id) {
    return await API.fetch(`movies?id=${id}`);
  },
  searchMovies: async function (query, order, genre) {
    return await API.fetch(
      `movies/search?query=${query}&order=${order}&genre=${genre}`,
    );
  },
  register: async function (name, email, password) {
    return await API.send("account/register", { name, email, password });
  },
  login: async function (email, password) {
    return await API.send("account/authenticate", { email, password });
  },
  getFavorites: async function () {
    return await API.fetch("account/favorites");
  },

  getWatchlist: async function () {
    return await API.fetch("account/watchlist");
  },
  saveToCollection: async function (movie_id, collection) {
    return await API.send("account/save-to-collection", {
      movie_id,
      collection,
    });
  },
  send: async function (serviceName, data) {
    try {
      const response = await fetch(API.baseURL + "/" + serviceName, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: app.Store.jwt ? `Bearer ${app.Store.jwt}` : null,
        },
        body: JSON.stringify(data),
      });
      const result = await response.json();
      return result;
    } catch (e) {
      console.error("Error in API send:", e);
    }
  },
  fetch: async function (serviceName) {
    try {
      const response = await fetch(API.baseURL + "/" + serviceName, {
        headers: {
          Authorization: app.Store.jwt ? `Bearer ${app.Store.jwt}` : null,
        },
      });
      const result = await response.json();
      return result;
    } catch (e) {
      console.error("Error in API fetch:", e);
    }
  },
};

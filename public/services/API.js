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

  fetch: async function (serviceName) {
    try {
      const response = await fetch(API.baseURL + "/" + serviceName);
      const result = await response.json();
      return result;
    } catch (e) {
      console.error("Error in API fetch:", e);
    }
  },
};

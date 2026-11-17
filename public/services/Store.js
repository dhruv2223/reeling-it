// The server issues JWTs that expire after 72h. An expired token must count
// as logged out, otherwise protected pages silently get 401s.
function isExpired(jwt) {
  try {
    const payload = JSON.parse(
      atob(jwt.split(".")[1].replace(/-/g, "+").replace(/_/g, "/")),
    );
    return typeof payload.exp === "number" && payload.exp * 1000 < Date.now();
  } catch (e) {
    return true; // malformed token
  }
}

const Store = {
  jwt: null,
  get loggedIn() {
    return this.jwt !== null && !isExpired(this.jwt);
  },
};
const savedJwt = localStorage.getItem("jwt");
if (savedJwt && !isExpired(savedJwt)) {
  Store.jwt = savedJwt;
} else {
  localStorage.removeItem("jwt");
}

const proxiedStore = new Proxy(Store, {
  set: (target, prop, value) => {
    if (prop == "jwt") {
      target[prop] = value;
      if (value == null) {
        localStorage.removeItem("jwt");
      } else {
        localStorage.setItem("jwt", value);
      }
    }
    return true;
  },
});

export default proxiedStore;

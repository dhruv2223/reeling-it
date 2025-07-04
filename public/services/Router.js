import { routes } from "./Routes.js";

export const Router = {
  init: function () {
    window.addEventListener("popstate", (event) => {
      Router.go(window.location.pathname, false);
    });
    document.querySelectorAll(".navlink").forEach((a) => {
      a.addEventListener("click", (event) => {
        event.preventDefault();
        let href = a.getAttribute("href");
        Router.go(href, true); // true to add to history
      });
    });
    Router.go(window.location.pathname);
  },
  go: function (pathname, addToHistory = true) {
    if (addToHistory) {
      window.history.pushState(null, "", pathname);
    }

    let pageElement = null;
    let routePath = pathname.includes("?") ? pathname.split("?")[0] : pathname;
    let needsLogin = false;
    for (let i = 0; i < routes.length; i++) {
      const route = routes[i];
      // String path match
      if (typeof route.path === "string" && route.path === routePath) {
        pageElement = new route.component();

        needsLogin = route.loggedIn === true;
        break;
        // RegExp path match
      } else if (route.path instanceof RegExp) {
        let match = route.path.exec(pathname); // not routePath if you want to capture query
        if (match) {
          pageElement = new route.component();
          pageElement.param = match.slice(1); // array of capture groups

          needsLogin = route.loggedIn === true;
          break;
        }
      }
    }
    if (needsLogin && app.Store.loggedIn == false) {
      console.log("i am here");
      app.Router.go("/account/login");
      return;
    }
    const main = document.querySelector("main");
    if (pageElement == null) {
      main.innerHTML = `<h1>Page not found</h1>`;
    } else {
      // If a page component was found
      function updatePage() {
        main.innerHTML = "";
        main.appendChild(pageElement);
      }
      const oldPage = document.querySelector("main").firstElementChild;
      pageElement.style.viewTransitionName = "new";
      if (oldPage) oldPage.style.viewTransitionName = "old";
      if (!document.startViewTransition) {
        updatePage();
      } else {
        document.startViewTransition(() => {
          updatePage();
        });
      }
    }
  },
};

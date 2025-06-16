import { CollectionPage } from "./CollectionPage.js";
import { API } from "../services/API.js";
export class WatchlistPage extends CollectionPage {
  constructor() {
    super(API.getWatchlist, "Watchlist");
  }
}
customElements.define("watchlist-page", WatchlistPage);

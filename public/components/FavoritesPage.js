import { CollectionPage } from "./CollectionPage.js";
import { API } from "../services/API.js";
export class FavoritesPage extends CollectionPage {
  constructor() {
    super(API.getFavorites, "Favorite Movies");
  }
}
customElements.define("favorite-page", FavoritesPage);

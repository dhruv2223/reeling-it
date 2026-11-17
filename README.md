# ReelingIt

A full-stack movie catalog: browse popular and random picks, search and filter about 5,000 movies, open a movie's details with its trailer and cast, and keep personal **Favorites** and **Watchlist** collections. Accounts support email and password login as well as **passkeys** (WebAuthn).

The backend is plain **Go** (standard-library `net/http`) on top of **PostgreSQL**. The frontend is a **single-page app written in vanilla JavaScript** with Web Components and no framework or build step.

---

## Contents

- [Features](#features)
- [Tech stack](#tech-stack)
- [Architecture](#architecture)
  - [System overview](#system-overview)
  - [Request lifecycle](#request-lifecycle)
  - [Authentication](#authentication)
  - [Data model](#data-model)
- [Project structure](#project-structure)
- [Setup](#setup)
- [Configuration](#configuration)
- [API reference](#api-reference)
- [Troubleshooting](#troubleshooting)

---

## Features

- **Home**: this week's top movies (by popularity) and a random selection.
- **Search**: by title or overview, with a genre filter and sorting by popularity, score, name or release date.
- **Movie details**: poster, embedded YouTube trailer, score, genres, overview and cast.
- **Accounts**: register and log in with email and password (bcrypt-hashed, JWT session).
- **Passkeys**: add a passkey from *My Account*, then log in without a password.
- **Collections**: add movies to your Favorites or Watchlist.
- **Client-side routing** with the History API and animated page transitions (View Transitions API).

## Tech stack

| Layer | Technology |
|---|---|
| Backend | Go 1.24, `net/http` |
| Database | PostgreSQL (driver: `lib/pq`) |
| Auth | JWT (`golang-jwt/jwt/v5`), bcrypt (`x/crypto`), WebAuthn (`go-webauthn/webauthn`) |
| Config | `.env` loaded with `joho/godotenv` |
| Frontend | HTML, CSS, vanilla JavaScript (ES modules), Web Components, `<template>` |
| Frontend libraries | [SimpleWebAuthn browser](https://simplewebauthn.dev/) 13.3.0 (passkeys only, loaded from a CDN) |
| Dev tooling | [Air](https://github.com/air-verse/air) for live reload (optional) |

---

## Architecture

### System overview

One Go binary does two jobs: it **serves the static SPA** from `public/`, and it **exposes a JSON REST API** under `/api`. Inside the server, each request moves through three layers: **handlers** (HTTP), **repositories** (SQL), then **models** (plain structs). Handlers only depend on repository *interfaces* (`data/interfaces.go`), so the storage layer can be swapped out.

```mermaid
%%{init: {"theme": "base", "themeVariables": {"fontFamily": "Inter, Segoe UI, Helvetica, Arial, sans-serif", "fontSize": "15px", "lineColor": "#64748b", "edgeLabelBackground": "#ffffff", "primaryTextColor": "#0f172a"}, "flowchart": {"curve": "basis", "nodeSpacing": 35, "rankSpacing": 55}}}%%
flowchart LR
    User(["👤 User"])

    subgraph Browser["🖥️ Browser · Single Page App"]
        direction TB
        HTML["📄 index.html<br/>HTML templates · styles.css"]
        App["⚙️ app.js<br/>global app controller"]
        Router["🧭 Router + Routes<br/>History API · View Transitions"]
        Pages["🧩 Web Components<br/>Home · Movies · Movie details<br/>Login · Register · Account · Collections"]
        Services["🔌 Services<br/>API.js · Passkey.js · Store.js"]
        LS[("💾 localStorage<br/>JWT")]
        HTML --> App --> Router --> Pages --> Services
        Services <--> LS
    end

    subgraph Server["🚀 Go server · :8080"]
        direction TB
        Mux{{"🚦 http.ServeMux"}}
        Static["📁 Static files<br/>public/ + SPA fallback"]
        MW["🛡️ AuthMiddleware<br/>JWT check"]
        H["🎛️ Handlers<br/>Movie · Account · WebAuthn"]
        R["🗄️ Repositories<br/>Movie · Account · Passkey"]
        M["📦 Models<br/>Movie · Genre · Actor · User · PasskeyUser"]
        Log["📝 Logger<br/>stdout · app.log"]
        Mux --> Static
        Mux --> H
        Mux --> MW --> H
        H --> R --> M
        H -.-> Log
        R -.-> Log
    end

    DB[("🐘 PostgreSQL<br/>movies · users · passkeys")]
    Ext["🎬 External media<br/>TMDB posters · YouTube trailers"]

    User -->|"opens the app"| HTML
    Static -.->|"serves HTML, JS, CSS"| HTML
    Services ==>|"JSON · /api/*"| Mux
    R ==>|"SQL over TCP"| DB
    Pages -.->|"img · iframe"| Ext

    classDef user fill:#1e293b,stroke:#0f172a,stroke-width:2px,color:#ffffff
    classDef ui fill:#fce7f3,stroke:#db2777,stroke-width:2px,color:#831843
    classDef js fill:#ecfccb,stroke:#65a30d,stroke-width:2px,color:#365314
    classDef store fill:#fef9c3,stroke:#ca8a04,stroke-width:2px,color:#713f12
    classDef gate fill:#ffedd5,stroke:#ea580c,stroke-width:2px,color:#7c2d12
    classDef go fill:#ccfbf1,stroke:#0d9488,stroke-width:2px,color:#134e4a
    classDef log fill:#f1f5f9,stroke:#94a3b8,stroke-width:1px,stroke-dasharray:4 3,color:#334155
    classDef db fill:#ede9fe,stroke:#7c3aed,stroke-width:3px,color:#4c1d95
    classDef ext fill:#fee2e2,stroke:#dc2626,stroke-width:2px,color:#7f1d1d

    class User user
    class HTML,Pages ui
    class App,Router,Services js
    class LS store
    class Mux,MW gate
    class Static,H,R,M go
    class Log log
    class DB db
    class Ext ext

    style Browser fill:#f7fee7,stroke:#84cc16,stroke-width:2px,color:#3f6212
    style Server fill:#f0fdfa,stroke:#14b8a6,stroke-width:2px,color:#115e59

    linkStyle default stroke:#64748b,stroke-width:2px
    linkStyle 14 stroke:#0d9488,stroke-width:2px
    linkStyle 15 stroke:#ea580c,stroke-width:3px
    linkStyle 16 stroke:#7c3aed,stroke-width:3px
    linkStyle 17 stroke:#dc2626,stroke-width:2px
```

**Colour key:** 🩷 markup and pages · 🟢 client-side JavaScript · 🟡 browser storage · 🟠 routing and the auth gate · 🩵 Go server layers · 🟣 database · 🔴 external media. Thick arrows are the two main data paths: the JSON API and SQL.

**Frontend.**
- `index.html` contains every page as an HTML `<template>`. Each page is a Web Component (a custom element) that clones its template and fills it with data from `API.js`.
- `Router.js` maps URLs to components and swaps the contents of `<main>`, using `history.pushState` for navigation.
- `Store.js` is a small `Proxy`-backed store that keeps the JWT in `localStorage`. It treats an expired token as logged out.

**Backend.**
- Requests for `/movies`, `/movies/*` and `/account/*` get `index.html`, so deep links and page reloads work. The client router takes over from there.
- Everything else under `/` is served from `public/`.

### Request lifecycle

What happens when a user opens a movie page directly, for example by pasting `/movies/234` into the address bar:

```mermaid
%%{init: {"theme": "base", "themeVariables": {"fontFamily": "Inter, Segoe UI, Helvetica, Arial, sans-serif", "actorBkg": "#ffffff", "actorBorder": "#334155", "actorTextColor": "#0f172a", "actorLineColor": "#94a3b8", "signalColor": "#334155", "signalTextColor": "#0f172a", "noteBkgColor": "#fef9c3", "noteBorderColor": "#ca8a04", "noteTextColor": "#713f12", "activationBkgColor": "#99f6e4", "activationBorderColor": "#0d9488", "sequenceNumberColor": "#ffffff"}}}%%
sequenceDiagram
    autonumber
    box rgb(252, 231, 243) 🖥️ Client
        actor U as 👤 User
        participant B as Browser (SPA)
    end
    box rgb(204, 251, 241) 🚀 Go server
        participant S as ServeMux
        participant H as MovieHandler
        participant R as MovieRepository
    end
    box rgb(237, 233, 254) 🐘 Database
        participant DB as PostgreSQL
    end

    rect rgb(255, 237, 213)
        Note over U,S: ① Load the app shell
        U->>B: open /movies/234
        B->>S: GET /movies/234
        S-->>B: index.html (SPA fallback)
        B->>B: Router matches /movies/:id → MovieDetailsPage
    end

    rect rgb(240, 253, 250)
        Note over B,DB: ② Fetch the data
        B->>S: GET /api/movies?id=234
        S->>+H: GetMovie
        H->>+R: GetMovieByID(234)
        R->>+DB: SELECT movie, genres, keywords, cast
        DB-->>-R: rows
        R-->>-H: models.Movie
        H-->>-B: 200 JSON (404 unknown · 400 bad id)
    end

    rect rgb(253, 242, 248)
        Note over U,B: ③ Render
        B->>B: fill the template: poster, trailer, metadata, cast
        B-->>U: 🎬 movie page
    end
```

Later navigation inside the app makes no new page requests. The router swaps components and only calls the API.

### Authentication

**Password login.**
- `POST /api/account/register` or `POST /api/account/authenticate` returns a signed **JWT**, valid for 72 hours. The client stores it and sends it as `Authorization: Bearer <jwt>`.
- `AuthMiddleware` checks the token on protected routes and puts the user's email into the request context.

**Passkeys (WebAuthn).**
- Registration and login are each two round trips: *begin* (the server creates a challenge) and *end* (the server verifies the browser's signed response).
- The short-lived WebAuthn session is kept in server memory under a random ID, sent as an `HttpOnly` cookie scoped to `/api/passkey`.
- Credentials are stored as JSON in the `passkeys` table. A successful passkey login returns the same kind of JWT as a password login.

```mermaid
%%{init: {"theme": "base", "themeVariables": {"fontFamily": "Inter, Segoe UI, Helvetica, Arial, sans-serif", "actorBkg": "#ffffff", "actorBorder": "#334155", "actorTextColor": "#0f172a", "actorLineColor": "#94a3b8", "signalColor": "#334155", "signalTextColor": "#0f172a", "noteBkgColor": "#fef9c3", "noteBorderColor": "#ca8a04", "noteTextColor": "#713f12", "sequenceNumberColor": "#ffffff"}}}%%
sequenceDiagram
    autonumber
    box rgb(252, 231, 243) 🖥️ Client
        participant B as Browser
        participant A as 🔑 Authenticator<br/>Touch ID · phone · security key
    end
    box rgb(204, 251, 241) 🚀 Go server
        participant S as WebAuthn handlers
    end
    box rgb(237, 233, 254) 🐘 Database
        participant DB as PostgreSQL
    end

    rect rgb(224, 242, 254)
        Note over B,DB: ➕ Add a passkey (logged in, request carries the JWT)
        B->>S: POST /api/passkey/registration-begin
        S->>DB: load user + existing credentials
        S-->>B: creation options + Set-Cookie sid
        B->>A: startRegistration()
        Note right of A: new key pair created,<br/>private key never leaves the device
        A-->>B: public key + attestation
        B->>S: POST /api/passkey/registration-end (cookie sid)
        S->>S: verify against the stored session
        S->>DB: INSERT credential into passkeys
        S-->>B: { success: true }
    end

    rect rgb(220, 252, 231)
        Note over B,DB: 🔓 Log in with a passkey
        B->>S: POST /api/passkey/authentication-begin { email }
        S->>DB: load user + credentials
        S-->>B: request options + Set-Cookie sid
        B->>A: startAuthentication()
        A-->>B: signed challenge (assertion)
        B->>S: POST /api/passkey/authentication-end (cookie sid)
        S->>S: verify signature + sign counter
        S->>DB: UPDATE credential (new sign counter)
        S-->>B: { success: true, jwt }
    end
```

### Data model

Movies are linked to genres, keywords and actors through join tables. Users save movies through `user_movies`, where `relation_type` is either `favorite` or `watchlist`. Passkey credentials belong to a user.

```mermaid
%%{init: {"theme": "base", "themeVariables": {"fontFamily": "Inter, Segoe UI, Helvetica, Arial, sans-serif", "primaryColor": "#ede9fe", "primaryBorderColor": "#7c3aed", "primaryTextColor": "#3b0764", "mainBkg": "#ede9fe", "nodeBorder": "#7c3aed", "attributeBackgroundColorOdd": "#ffffff", "attributeBackgroundColorEven": "#f5f3ff", "lineColor": "#7c3aed", "tertiaryColor": "#ffffff", "edgeLabelBackground": "#ffffff", "textColor": "#3b0764"}}}%%
erDiagram
    users ||--o{ user_movies : saves
    movies ||--o{ user_movies : "saved in"
    users ||--o{ passkeys : owns
    movies ||--o{ movie_genres : has
    genres ||--o{ movie_genres : tags
    movies ||--o{ movie_keywords : has
    keywords ||--o{ movie_keywords : tags
    movies ||--o{ movie_cast : features
    actors ||--o{ movie_cast : "appears in"

    users {
        int id PK
        text name
        text email
        text password_hashed
        timestamp last_login
        timestamp time_created
        timestamp time_deleted
    }
    movies {
        int id PK
        int tmdb_id
        text title
        text tagline
        int release_year
        text overview
        float score
        float popularity
        text language
        text poster_url
        text trailer_url
    }
    genres {
        int id PK
        text name
    }
    keywords {
        int id PK
        text word
    }
    actors {
        int id PK
        text first_name
        text last_name
        text image_url
    }
    user_movies {
        int user_id PK, FK
        int movie_id PK, FK
        text relation_type PK "favorite or watchlist"
        timestamp time_added
    }
    movie_genres {
        int movie_id PK, FK
        int genre_id PK, FK
    }
    movie_keywords {
        int movie_id PK, FK
        int keyword_id PK, FK
    }
    movie_cast {
        int movie_id PK, FK
        int actor_id PK, FK
    }
    passkeys {
        int id PK
        int user_id
        text keys "credential JSON"
    }
```

---

## Project structure

```
.
├── main.go                  # wiring: env, DB, WebAuthn, routes, server
├── .air.toml                # live-reload config (optional)
├── data/                    # repositories (SQL) + storage interfaces
│   ├── interfaces.go        #   MovieStorage, AccountStorage, PasskeyStore
│   ├── movie_repository.go
│   ├── account_repository.go
│   └── passkey_repository.go
├── handlers/                # HTTP handlers
│   ├── movie_handlers.go
│   ├── account_handlers.go  #   includes AuthMiddleware
│   └── passkey_handlers.go
├── models/                  # Movie, Genre, Actor, User, PasskeyUser
├── token/                   # JWT creation + secret lookup
├── logger/                  # info → stdout, errors → app.log
├── import/
│   └── database-dump.sql    # schema + movie dataset (load once, see Setup)
└── public/                  # the SPA, served as-is
    ├── index.html           #   shell + all page <template>s
    ├── styles.css
    ├── app.js               #   global `app` controller
    ├── app.webmanifest
    ├── components/          #   Web Components (one per page/widget)
    ├── services/            #   API, Router, Routes, Store, Passkey
    └── images/
```

---

## Setup

### Prerequisites

- **Go 1.24+** (`go version`)
- **PostgreSQL 14+**: a local install, Docker, or a hosted database
- A modern browser. Passkeys need Chrome, Edge, Safari or Firefox with a platform authenticator, a phone, or a security key.
- Optional: [Air](https://github.com/air-verse/air) for live reload: `go install github.com/air-verse/air@latest`

### 1. Get the code

```bash
git clone <repository-url> reeling-it
cd reeling-it
go mod download
```

### 2. Create the database

The schema and movie data ship with the repository in `import/database-dump.sql` (about 8 MB). It holds roughly 4,800 movies with genres, keywords and cast, derived from [TMDB](https://www.themoviedb.org/), plus empty `users` and `user_movies` tables. Poster images and trailers are loaded live from TMDB and YouTube.

> **Warning:** the dump starts with `DROP TABLE IF EXISTS` for every table. Load it **only once**: running it again deletes all registered users and their favorites and watchlists.

**Option A: Docker** (quickest)

```bash
docker run -d --name reelingit-db \
  -e POSTGRES_PASSWORD=dev -e POSTGRES_DB=reelingit \
  -p 5432:5432 postgres:16-alpine
```

Wait a few seconds for it to start. Then:

```bash
docker exec -i reelingit-db psql -U postgres -d reelingit < import/database-dump.sql
```

**Option B: an existing PostgreSQL server or hosted database**

```bash
createdb reelingit            # skip for a hosted DB that already has a database
psql "<your-connection-url>" -f import/database-dump.sql
```

### 3. Create the `passkeys` table

The dump doesn't include it. Run this once against the same database (with Docker: `docker exec -i reelingit-db psql -U postgres -d reelingit`):

```sql
CREATE SEQUENCE IF NOT EXISTS passkeys_id_seq;

CREATE TABLE IF NOT EXISTS "public"."passkeys" (
    "id" int4 NOT NULL DEFAULT nextval('passkeys_id_seq'::regclass),
    "user_id" int4,
    "keys" text,
    PRIMARY KEY ("id")
);
```

Check the import worked: `SELECT count(*) FROM movies;` should return **4797**.

### 4. Configure environment variables

Create a `.env` file in the project root. It's gitignored, and **the server won't start without it**:

```dotenv
DATABASE_URL=postgres://postgres:dev@localhost:5432/reelingit?sslmode=disable
JWT_SECRET=replace-with-a-long-random-string
```

Generate a secret with `openssl rand -base64 48`. For hosted databases that need TLS, use `sslmode=require` instead.

### 5. Run

```bash
go run .
# or, with live reload:
air
```

Open **<http://localhost:8080>**. Use `localhost`, not `127.0.0.1`: passkeys are bound to that exact origin (see [Configuration](#configuration)).

The server checks the database connection at startup and exits with a clear error if it can't reach it.

### 6. Try it out

1. Search for a movie, open it, and watch the trailer.
2. **Register** an account. You'll land on *My Account*.
3. Add a movie to **Favorites** or **Watchlist**.
4. On *My Account*, click **Add a Passkey for faster login** and follow your browser's prompt.
5. **Log out**. On the login page, type your email and click **Log In with a Passkey**.

### Build a binary

```bash
go build -o reelingit .
./reelingit        # run it from the project root: it reads ./.env and serves ./public
```

---

## Configuration

| Setting | Where | Notes |
|---|---|---|
| `DATABASE_URL` | `.env` / environment | **Required.** PostgreSQL connection URL. A variable already exported in your shell takes precedence over `.env`. |
| `JWT_SECRET` | `.env` / environment | Secret used to sign JWTs. If it's missing, a built-in development default is used, so **always set it** outside local development. |
| Port | `main.go` (`addr`) | `:8080` |
| Passkey relying party | `main.go` (`webauthn.Config`) | `RPID: "localhost"`, `RPOrigins: ["http://localhost:8080"]`. To deploy under a real domain, set these to that domain and its HTTPS origin, or passkeys will be rejected. |
| Logs | `app.log` | Errors go to `app.log` in the working directory; info goes to stdout. |

---

## API reference

All responses are JSON unless the status is an error, in which case the body is a plain-text message. Routes marked 🔒 need `Authorization: Bearer <jwt>`.

### Movies

| Method | Path | Description |
|---|---|---|
| GET | `/api/movies/top` | 20 most popular movies |
| GET | `/api/movies/random` | 20 random movies |
| GET | `/api/movies?id={id}` | One movie with genres, keywords and cast. Returns `400` for a bad ID, `404` if it doesn't exist. |
| GET | `/api/movies/search?query={q}&order={order}&genre={genreId}` | Searches title and overview. `order`: `popularity` (default), `score`, `name`, `date`. `genre` is optional. |
| GET | `/api/genres` | All genres |

### Account

| Method | Path | Body | Description |
|---|---|---|---|
| POST | `/api/account/register` | `{ name, email, password }` | Creates the account and returns `{ success, message, jwt }`. Returns `401` if the email is already taken. |
| POST | `/api/account/authenticate` | `{ email, password }` | Returns `{ success, message, jwt }`. Returns `401` if the credentials are wrong. |
| GET 🔒 | `/api/account/favorites` | – | The user's favorite movies |
| GET 🔒 | `/api/account/watchlist` | – | The user's watchlist |
| POST 🔒 | `/api/account/save-to-collection` | `{ movie_id, collection }` | `collection` is `favorite` or `watchlist` |

### Passkeys

| Method | Path | Body | Description |
|---|---|---|---|
| POST 🔒 | `/api/passkey/registration-begin` | – | Returns WebAuthn creation options and sets the `sid` cookie |
| POST 🔒 | `/api/passkey/registration-end` | attestation from the browser | Verifies the response and stores the credential |
| POST | `/api/passkey/authentication-begin` | `{ email }` | Returns request options and sets the `sid` cookie. Returns `404` if there's no account or no passkey. |
| POST | `/api/passkey/authentication-end` | assertion from the browser | Verifies the response and returns `{ success, jwt }` |

---

## Troubleshooting

| Symptom | Fix |
|---|---|
| `Error loading .env file` at startup | Create `.env` in the directory you run the server from (step 4). |
| `Error reaching the database` at startup | Check `DATABASE_URL`, and that PostgreSQL is running and reachable from this machine. |
| API returns 500, or the home page is empty | The dump probably wasn't imported into the database `DATABASE_URL` points at. Re-run step 2 and check `SELECT count(*) FROM movies;`. |
| Passkey errors mentioning the origin or RP ID | Open the app at `http://localhost:8080` exactly, or update `webauthn.Config` in `main.go` for your domain. |
| "No passkey is registered for this account" | Log in with your password first and add a passkey from *My Account*. |
| Passkey actions fail with a `relation "passkeys" does not exist` error in `app.log` | Run the SQL in step 3. |
| Logged out unexpectedly | Sessions (JWTs) last 72 hours. Log in again. |

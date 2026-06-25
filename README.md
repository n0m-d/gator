# Gator

A terminal RSS feed aggregator written in Go. Gator runs as an interactive TUI for browsing posts, managing feeds, and collecting articles in the background. A small set of CLI commands remains for account setup and scripting.

> https://www.boot.dev/courses/build-blog-aggregator-golang

## Features

- Interactive terminal UI (Bubble Tea + Lip Gloss) for day-to-day use
- Built-in login and registration screen on first launch
- Browse paginated posts from feeds you follow
- Add, follow, and unfollow feeds from the TUI
- Background aggregation with progress bar (press `r` to start/stop)
- Copy post URLs to the clipboard
- PostgreSQL storage with goose migrations and sqlc-generated queries
- Minimal CLI for `login`, `register`, `users`, and `reset`

## Prerequisites

- [Go](https://go.dev/) 1.25+
- [PostgreSQL](https://www.postgresql.org/)
- [goose](https://github.com/pressly/goose) for database migrations
- [sqlc](https://sqlc.dev/) for generating type-safe database code

Install tooling:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## Setup

### 1. Create a database

```bash
createdb gator
```

### 2. Configure the app

Create `~/.gatorconfig.json` in your home directory:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Replace the connection string with your PostgreSQL credentials. `current_user_name` is set when you log in or register.

### 3. Run migrations

From the project root:

```bash
goose -dir sql/schema postgres "$DB_URL" up
```

Or pass the connection string directly:

```bash
goose -dir sql/schema postgres "postgres://username:password@localhost:5432/gator?sslmode=disable" up
```

### 4. Build and run

```bash
go build -o gator ./cmd/gator
./gator
```

On first launch you will see a login/register screen. You can also create an account from the shell:

```bash
./gator register alice
./gator
```

## TUI usage

Launch with no arguments:

```bash
./gator
```

### Authentication screen

Shown when no user is logged in.

| Key | Action |
|-----|--------|
| `tab` | Switch between Login and Register |
| `enter` | Submit |
| `q` / `ctrl+c` | Quit |

### Main app

| Key | Action |
|-----|--------|
| `tab` / `1` / `2` | Switch between Posts and Following tabs |
| `j` / `k` or `↑` / `↓` | Navigate list |
| `h` / `l` or `←` / `→` | Previous / next page (Posts tab) |
| `v` | Copy selected post URL to clipboard |
| `u` | Unfollow selected feed (Following tab) |
| `a` | Add a feed (Following tab) |
| `r` | Start/stop background aggregation |
| `q` / `ctrl+c` | Quit |

While aggregation is running, a progress bar shows fetch status and time until the next scrape. The status bar label switches from **User** to **Agg**.

### Example workflow

```bash
./gator register alice
./gator
```

In the TUI:

1. Switch to the **Following** tab (`2`)
2. Press `a` to add a feed (name + URL)
3. Press `r` to start collecting posts
4. Switch to **Posts** (`1`) to browse articles

## CLI commands

The TUI is the primary interface. These commands are available for scripts, automation, or quick account management:

| Command | Description |
|---------|-------------|
| `register <name>` | Create a new user and log in |
| `login <name>` | Log in as an existing user |
| `users` | List all users (current user marked with `*`) |
| `reset` | Delete all users (development) |
| `help` | Show available CLI commands |

```bash
./gator help
./gator login alice
./gator users
```

Feed management, browsing, and aggregation are only available in the TUI.

## Project structure

```
.
├── cmd/gator/
│   └── main.go                 # Entry point: TUI by default, CLI for auth
├── internal/
│   ├── cli/                    # Login, register, users, reset
│   │   ├── cli.go
│   │   └── users.go
│   ├── config/
│   │   └── config.go           # Reads/writes ~/.gatorconfig.json
│   ├── database/               # sqlc-generated query code (do not edit)
│   ├── rss/
│   │   └── rss.go              # RSS fetch and parse
│   ├── scraper/
│   │   └── scraper.go          # Feed scraping and post persistence
│   └── tui/
│       ├── model.go            # Bubble Tea model and key handling
│       ├── view.go             # Lip Gloss rendering
│       ├── styles.go           # Theme
│       ├── auth.go             # Login/register screen
│       ├── add_feed.go         # Add feed dialog
│       ├── agg.go              # Background aggregation
│       └── terminal.go         # Terminal cleanup on exit
├── sql/
│   ├── schema/                 # Goose migration files
│   └── queries/                # sqlc query definitions
├── sqlc.yaml
├── go.mod
└── go.sum
```

## Database schema

| Table | Purpose |
|-------|---------|
| `users` | Registered users |
| `feeds` | RSS feed metadata (name, URL, owner) |
| `feed_follows` | Many-to-many relationship between users and feeds |
| `posts` | Individual articles scraped from feeds |

Posts are deduplicated by URL. If a post with the same URL already exists, the scraper silently skips it.

## Development

### Regenerate database code

After changing files in `sql/queries/` or `sql/schema/`:

```bash
sqlc generate
```

### Run tests

```bash
go test ./...
```

### Add a new migration

Create a numbered file in `sql/schema/` (e.g. `006_something.sql`):

```sql
-- +goose Up
-- your migration

-- +goose Down
-- rollback
```

Apply it:

```bash
goose -dir sql/schema postgres "$DB_URL" up
```

## License

See [LICENSE](LICENSE).

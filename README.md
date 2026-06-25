# Gator

A command-line RSS feed aggregator written in Go. Gator lets you register users, subscribe to feeds, periodically scrape new posts, and browse saved articles from the terminal.

> https://www.boot.dev/courses/build-blog-aggregator-golang

## Features

- User registration and session management via a local config file
- Add and list RSS feeds
- Follow and unfollow feeds per user
- Background aggregation that fetches feeds on a schedule and stores posts in PostgreSQL
- Browse recent posts from feeds you follow

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

### 2. Configure the CLI

Create `~/.gatorconfig.json` in your home directory:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Replace the connection string with your PostgreSQL credentials. `current_user_name` is set automatically when you register or log in.

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
go build -o gator .
./gator <command> [args...]
```

You can also run without building:

```bash
go run . <command> [args...]
```

## Usage

### User commands

| Command | Description |
|---------|-------------|
| `register <name>` | Create a new user and log in as them |
| `login <name>` | Switch to an existing user |
| `users` | List all users (current user is marked) |
| `reset` | Delete all users |

### Feed commands

| Command | Description |
|---------|-------------|
| `addfeed <title> <url>` | Add a feed and automatically follow it (requires login) |
| `feeds` | List all feeds |
| `follow <url>` | Follow an existing feed by URL (requires login) |
| `following` | List feeds the current user follows (requires login) |
| `unfollow <url>` | Stop following a feed (requires login) |

### Aggregation and posts

| Command | Description |
|---------|-------------|
| `agg <duration>` | Continuously fetch feeds on an interval (e.g. `30s`, `1m`, `5m`) |
| `browse [limit]` | Show recent posts from followed feeds (requires login; default limit is 2) |

### Example workflow

```bash
./gator register alice
./gator addfeed "Hacker News" https://news.ycombinator.com/rss
./gator agg 1m
```

In another terminal:

```bash
./gator login alice
./gator browse 5
```

## Project structure

```
.
├── main.go                 # Entry point, command registration
├── commands.go             # Command dispatcher
├── middleware.go           # Login-required middleware
├── user_handler.go         # User commands (register, login, users, reset)
├── rss_handler.go          # Feed commands (addfeed, feeds, agg)
├── feed_follow_handler.go  # Follow/unfollow commands
├── posts_handler.go        # Browse command
├── aggregator.go           # Feed scraping and post persistence
├── rss.go                  # RSS fetching and date parsing
├── internal/
│   ├── config/
│   │   └── config.go       # Reads/writes ~/.gatorconfig.json
│   └── database/           # sqlc-generated query code (do not edit by hand)
├── sql/
│   ├── schema/             # Goose migration files
│   │   ├── 001_users.sql
│   │   ├── 002_feeds.sql
│   │   ├── 003_feed_follows.sql
│   │   ├── 004_feeds_last_fetched_at.sql
│   │   └── 005_posts.sql
│   └── queries/            # sqlc query definitions
│       ├── users.sql
│       ├── feeds.sql
│       ├── feed_follows.sql
│       └── posts.sql
├── sqlc.yaml               # sqlc configuration
├── go.mod
└── go.sum
```

## Database schema

| Table | Purpose |
|-------|---------|
| `users` | Registered CLI users |
| `feeds` | RSS feed metadata (name, URL, owner) |
| `feed_follows` | Many-to-many relationship between users and feeds |
| `posts` | Individual articles scraped from feeds |

Posts are deduplicated by URL. If a post with the same URL already exists, the scraper silently skips it.

## Development

### Regenerate database code

After changing files in `sql/queries/` or `sql/schema/`, run:

```bash
sqlc generate
```

Generated Go code is written to `internal/database/`.

### Add a new migration

Create a new numbered file in `sql/schema/` (e.g. `006_something.sql`) using goose directives:

```sql
-- +goose Up
-- your migration

-- +goose Down
-- rollback
```

Then apply it:

```bash
goose -dir sql/schema postgres "$DB_URL" up
```

### Add a new SQL query

1. Add the query to the appropriate file in `sql/queries/`
2. Run `sqlc generate`
3. Use the generated method from `internal/database` in a handler

## License

See [LICENSE](LICENSE).

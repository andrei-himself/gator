# gator

**gator** is a lightweight CLI blog aggregator written in Go.  
It connects to a PostgreSQL database, lets you register users, follow RSS feeds, and browse fetched posts — all from your terminal.

---

## Requirements

You’ll need:

- **Go** (v1.20 or newer) → [https://go.dev/dl/](https://go.dev/dl/)  
- **PostgreSQL** → [https://www.postgresql.org/download/](https://www.postgresql.org/download/)

Make sure PostgreSQL is running and you can connect via `psql`.

---

## Installation

Install the CLI directly with:

```bash
go install github.com/andrei-himself/gator@latest
```

Ensure your Go bin directory is in your `PATH` (usually `~/go/bin`):

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Then check it works:

```bash
gator
```

---

## Configuration

`gator` uses a small config file to store your **database URL** and the **current user**.

Create a file at:

```bash
~/.gatorconfig.json
```

Example contents:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

- `db_url`: your PostgreSQL connection string  
- `current_user_name`: set automatically when you log in  

---

## Commands Overview

Run commands as:

```bash
gator <command> [args...]
```

| Command | Description |
|----------|-------------|
| `register <username>` | Create a new user and set it as current |
| `login <username>` | Set an existing user as current |
| `users` | List all users (marks the current one) |
| `reset` | Delete all users, feeds, and follows |
| `addfeed <name> <url>` | Add a new feed (auto-follows it) |
| `feeds` | List all feeds with owners |
| `follow <url>` | Follow a feed by URL |
| `unfollow <url>` | Unfollow a feed |
| `following` | Show feeds followed by current user |
| `agg <duration>` | Continuously fetch feeds every given duration (e.g. `1m`) |
| `browse [limit]` | Show recent posts for followed feeds (default limit = 2) |

---

## Example Usage

```bash
# Create and log in as a user
gator register alice
gator login alice

# Add and follow a feed
gator addfeed "Boot.dev Blog" https://blog.boot.dev/index.xml

# Fetch feeds every minute
gator agg 1m

# Browse latest posts
gator browse 5

# Unfollow a feed
gator unfollow https://blog.boot.dev/index.xml

# List users and feeds
gator users
gator feeds
```

---

## Database Setup (quick example)

```bash
createdb gator
createuser gatoruser --pwprompt
```

Update your `~/.gatorconfig.json` with the correct `db_url`:

```
postgres://gatoruser:yourpassword@localhost:5432/gator?sslmode=disable
```

---

## Troubleshooting

- **DB errors:** check your `db_url` and that PostgreSQL is running.  
- **Command not found:** ensure `~/go/bin` is in your PATH.  
- **Help:** run `gator` with no args to see available commands.

---

## License

MIT License © 2025 [andrei-himself](https://github.com/andrei-himself)

---

**Enjoy using gator!**

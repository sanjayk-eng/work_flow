# sanjay-khandelwal — Go REST API

A modular Go REST API built with Gin, PostgreSQL, and JWT authentication.

---

## Tech Stack

- **Language:** Go 1.26
- **Framework:** [Gin](https://github.com/gin-gonic/gin)
- **Database:** PostgreSQL via [pgx/v5](https://github.com/jackc/pgx)
- **Migrations:** [Goose v3](https://github.com/pressly/goose)
- **Auth:** JWT (access + refresh token pattern)
- **Config:** Environment variables via [godotenv](https://github.com/joho/godotenv)

---

## Project Structure

```
.
├── cmd/
│   └── api/
│       ├── main.go                          # Entry point — loads config, DB, migrations, server
│       └── .env                             # Local environment variables
│
├── internal/
│   ├── middleware/
│   │   └── auth.go                          # JWT auth middleware (stub)
│   │
│   ├── modules/                             # Feature modules (handler / service / repository pattern)
│   │   ├── auth/
│   │   │   ├── dto/
│   │   │   │   ├── request.go               # Auth request DTOs
│   │   │   │   └── response.go              # Auth response DTOs
│   │   │   ├── handler.go                   # HTTP handlers
│   │   │   ├── models.go                    # Domain models
│   │   │   ├── module.go                    # Dependency wiring
│   │   │   ├── repository.go                # Database queries
│   │   │   ├── router.go                    # Route registration
│   │   │   └── service.go                   # Business logic
│   │   ├── organization/                    # (placeholder)
│   │   └── user/                            # (placeholder)
│   │
│   └── shared/
│       ├── core/                            # Core infrastructure (config, DB, server)
│       │   ├── config/
│       │   │   ├── config.go                # Config, AppConfig, DBConfig, RedisConfig structs
│       │   │   ├── const.go                 # Env key constants
│       │   │   ├── env.go                   # godotenv loader
│       │   │   ├── loader.go                # NewConfig() — reads env into structs
│       │   │   └── provider.go              # Get() singleton via sync.Once
│       │   │
│       │   ├── database/
│       │   │   └── postgres/
│       │   │       ├── connections.go       # pgxpool setup, New(), Close()
│       │   │       ├── const.go             # Pool defaults (max/min conns, timeouts)
│       │   │       ├── dsn.go               # BuildDSN() — postgres connection string
│       │   │       └── migration.go         # MigrateUp(), MigrateDown() via Goose
│       │   │
│       │   └── server/
│       │       └── server.go                # Gin engine setup, New(), Run()
│       │
│       └── security/
│           ├── jwt/
│           │   └── config.go                # JWT config struct (secret, expiry, issuer)
│           └── password/
│               ├── argon2.go                # Pure generateHash() using argon2id
│               ├── config.go                # Argon2 params (memory, iterations, etc.)
│               └── password.go              # Hasher — Hash() and Verify()
│
├── migrations/                              # SQL migration files (Goose)
│   ├── 20260605155734_tbl_user.sql          # users table + user_status enum
│   ├── 20260605155848_create_profiles.sql   # profiles table
│   ├── 20260605160022_create_email_verifications.sql  # email_verifications table
│   └── 20260605160339_create_sessions.sql   # sessions table
│
├── pkg/                                     # Shared/reusable packages (future use)
├── go.mod
├── go.sum
└── README.md
```

---

## Database Schema

| Table                 | Description                                      |
|-----------------------|--------------------------------------------------|
| `users`               | Core user record (email, password hash, status)  |
| `profiles`            | User profile (first name, last name, avatar)     |
| `email_verifications` | Email verification tokens with expiry            |
| `sessions`            | Refresh token sessions with revocation support   |

---

## Configuration

Copy `.env` from `cmd/api/.env` and fill in your values:

```env
# App
APP_PORT=8082
APP_ENV=dev

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=yourpassword
DB_NAME=work_flow
SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0

# Logging
LOG_LEVEL=info
LOG_MODE=json
```

All env keys and their defaults are defined in `internal/shared/config/const.go`.

---

## Getting Started

### Prerequisites

- Go 1.26+
- PostgreSQL running locally (or via Docker)
- Redis (optional, for future session/cache use)

### Run

```bash
# From the project root
cd cmd/api
go run main.go
```

The server starts on the port defined by `APP_PORT` (default `8080`).

Migrations run automatically on startup via Goose pointing at `../../migrations`.

### Run Migrations Manually

Migrations are handled automatically on startup. To run them manually you can use the Goose CLI:

```bash
goose -dir migrations postgres "postgres://user:pass@localhost:5432/work_flow?sslmode=disable" up
goose -dir migrations postgres "postgres://user:pass@localhost:5432/work_flow?sslmode=disable" down
```

---

## Architecture

Each feature lives in `internal/moduler/<feature>/` and follows a layered pattern:

```
router.go      → registers Gin routes
handler.go     → parses requests, calls service, writes responses
service.go     → business logic
repository.go  → database queries
models.go      → domain structs
module.go      → wires the layers together (dependency injection)
dto/           → request and response types
```

The `internal/shared/` layer provides cross-cutting infrastructure (config, DB, JWT, server) that modules consume but don't own.

---

## Security Notes

- JWT secret defaults to `change-this-secret` — always override via `JWT_SECRET` env in production.
- Passwords are stored as hashes (see `internal/shared/security/password`).
- Refresh tokens are stored as hashes in the `sessions` table, never in plain text.
- SSL mode defaults to `disable` for local dev — set `SSL_MODE=require` in production.

---

## Shared Layer Evaluation

### config/ ✅ Good

Well structured across 5 files — types, constants, env loader, loader, and singleton provider. `sync.Once` pattern is correct for config initialization.

> Note: `godotenv.Load()` looks for `.env` in the working directory. Since the binary runs from `cmd/api/`, this works — but only when run from that directory.

---

### database/postgres/ ✅ Good

Pool setup is clean with sensible defaults in `const.go`. DSN builder is isolated. Migration helpers are correct.

Issues to address:

- `init()` in `migration.go` calls `panic` on failure — hard to test and unrecoverable. Move `goose.SetDialect` into `MigrateUp`/`MigrateDown` or call it once in `main`.
- `MigrateDown` is exported with no guard — one wrong call drops the schema. Consider unexporting it or adding an explicit confirmation mechanism.

---

### security/password/ ✅ Excellent

- Argon2id algorithm — correct choice, stronger than bcrypt for offline attacks
- `subtle.ConstantTimeCompare` — prevents timing attacks
- `crypto/rand` for salt — cryptographically secure
- PHC string format (`$argon2id$v=19$...`) — industry standard encoding
- `generateHash` is a pure function with no IO side effects
- `config` and `defaultConfig()` are unexported — correct encapsulation

Future improvement: add a `NeedsRehash(encoded string) bool` function for when cost parameters are rotated.

---

### security/jwt/ ⚠️ Incomplete

Only a config struct exists. Missing:

- Token generation (`Generate(claims) (string, error)`)
- Token parsing and validation (`Verify(token string) (claims, error)`)
- A `Claims` struct
- Any exported type — currently everything is unexported so nothing outside the package can use it
- Secret key is hardcoded as `"change-this-secret"` with no injection point from env

Needs a `New(secretKey []byte) *Service` constructor pattern, same as the password package.

---

### server/ ⚠️ Incomplete

Two issues:

**No recovery middleware** — `gin.New()` is used with zero middleware. A single unhandled panic crashes the server. Add at minimum:
```go
engine.Use(gin.Recovery())
```

**DB is never injected** — `server.New()` takes no arguments, so route handlers have no path to the database pool. `*postgres.DB` needs to be passed into `New()` or injected when modules are wired up.

---

### Shared Layer Summary

| Package              | Status        | Key Issue                                      |
|----------------------|---------------|------------------------------------------------|
| `config/`            | ✅ Good        | None                                           |
| `database/postgres/` | ✅ Good        | `init()` panic, `MigrateDown` unguarded        |
| `security/password/` | ✅ Excellent   | None                                           |
| `security/jwt/`      | ⚠️ Incomplete  | No token logic, secret not injectable from env |
| `server/`            | ⚠️ Incomplete  | No recovery middleware, DB not wired in        |

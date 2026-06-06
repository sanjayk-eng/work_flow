# sanjay-khandelwal — Go REST API

A modular Go REST API built with Gin, PostgreSQL, and JWT authentication.

---

## Tech Stack

- **Language:** Go 1.26
- **Framework:** [Gin](https://github.com/gin-gonic/gin)
- **Database:** PostgreSQL via [pgx/v5](https://github.com/jackc/pgx)
- **Migrations:** [Goose v3](https://github.com/pressly/goose)
- **Auth:** JWT access tokens (identity-only) + refresh tokens via DB sessions
- **Password hashing:** Argon2id via [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto)
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
│   │   ├── auth.go                          # Authenticate — JWT verify, sets user_id + email
│   │   ├── org.go                           # OrgContext — DB membership lookup, sets org_id + role
│   │   └── rbac.go                          # RequireRole — role-based access check
│   │
│   ├── moduler/                             # Feature modules (handler / service / repository pattern)
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
│       ├── core/
│       │   ├── config/
│       │   │   ├── config.go                # Config, AppConfig, DBConfig, RedisConfig, JWTConfig, LogConfig structs
│       │   │   ├── const.go                 # Env key constants (app, db, redis, jwt, log)
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
│       │   ├── logger/
│       │   │   └── logger.go                # slog JSON logger, WithContext(), FromContext()
│       │   │
│       │   ├── session/
│       │   │   ├── session.go               # Session struct, Repository interface, GenerateRefreshToken()
│       │   │   └── service.go               # Issue(), Rotate() with reuse detection, Revoke(), RevokeAll()
│       │   │
│       │   └── server/
│       │       └── server.go                # Gin engine, Recovery middleware, request logger injection
│       │
│       ├── apperr/
│       │   └── apperr.go                    # Typed domain errors — NotFound, Unauthorized, Conflict, etc.
│       │
│       └── security/
│           ├── jwt/
│           │   ├── claims.go                # Claims struct — user_id + email only
│           │   ├── config.go                # JWT config — secret key, expiry, issuer
│           │   └── service.go               # New(), CreateToken(), VerifyToken()
│           └── password/
│               ├── argon2.go                # Pure generateHash() using argon2id
│               ├── config.go                # Argon2 params (memory, iterations, parallelism)
│               └── password.go              # Hasher — Hash() and Verify()
│
├── migrations/
│   ├── 20260605155734_tbl_user.sql          # users table + user_status enum
│   ├── 20260605155848_create_profiles.sql   # profiles table
│   ├── 20260605160022_create_email_verifications.sql  # email_verifications table
│   └── 20260605160339_create_sessions.sql   # sessions table (refresh token store)
│
├── pkg/
│   ├── response/
│   │   └── json.go                          # Success(), Created(), Error(), Paginated(), PaginationMeta
│   ├── validator/
│   │   └── validator.go                     # Singleton validator, Validate(), strong_password custom rule
│   └── pagination/
│       └── pagination.go                    # ParsePage(), ParseCursor(), TotalPages(), Page, Cursor
│
├── go.mod
├── go.sum
└── README.md
```

---

## Auth Architecture

### JWT — identity only

Access tokens contain the minimum identity data needed. No roles, no org context, no session IDs.

```
Claims {
    user_id   string
    email     string
    iss, iat, exp
}
```

```go
svc := jwt.New([]byte(cfg.JWT.SecretKey))

token, err := svc.CreateToken(userID, email)   // issues 15-min access token
claims, err := svc.VerifyToken(tokenStr)        // returns Claims or sentinel error
```

Refresh tokens are random values stored as hashes in the `sessions` table — never embedded in JWTs.

### 3-layer middleware chain

```
Request
  ↓
Authenticate(jwtSvc)          → verifies JWT → sets user_id, email in context
  ↓
OrgContext(orgRepo)            → reads X-Org-ID header → DB membership lookup → sets org_id, role
  ↓
RequireRole("admin", "member") → checks role from context → 403 if not allowed
  ↓
Handler
```

### Router usage

```go
// any authenticated user
r.GET("/me", middleware.Authenticate(jwtSvc), handler.GetProfile)

// org-scoped route (any role)
org := r.Group("/org",
    middleware.Authenticate(jwtSvc),
    middleware.OrgContext(orgRepo),
)
org.GET("/data", handler.GetOrgData)

// admin only
org.DELETE("/member", middleware.RequireRole("admin"), handler.RemoveMember)

// admin or member
org.GET("/settings", middleware.RequireRole("admin", "member"), handler.GetSettings)
```

### Context keys

| Key        | Set by          | Value              |
|------------|-----------------|--------------------|
| `user_id`  | `Authenticate`  | UUID string        |
| `email`    | `Authenticate`  | user email         |
| `org_id`   | `OrgContext`    | UUID string        |
| `role`     | `OrgContext`    | e.g. admin, member |

---

## Database Schema

| Table                  | Description                                           |
|------------------------|-------------------------------------------------------|
| `users`                | Core user record — email, password hash, status       |
| `profiles`             | User profile — first name, last name, avatar          |
| `email_verifications`  | Email verification tokens with expiry                 |
| `sessions`             | Refresh token hashes — expires_at + revoked_at        |

---

## Configuration

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

# JWT
JWT_SECRET=your-strong-secret-here
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
JWT_ISSUER=app-service

# Logging
LOG_LEVEL=info
```

All keys and defaults are in `internal/shared/core/config/const.go`.

---

## Getting Started

### Prerequisites

- Go 1.26+
- PostgreSQL running locally (or Docker)

### Run

```bash
cd cmd/api
go run main.go
```

Server starts on `APP_PORT` (default `8080`). Migrations run automatically on startup.

### Migrations manually

```bash
goose -dir migrations postgres "postgres://user:pass@localhost:5432/work_flow?sslmode=disable" up
goose -dir migrations postgres "postgres://user:pass@localhost:5432/work_flow?sslmode=disable" down
```

---

## Module Pattern

Each feature in `internal/moduler/<feature>/` follows:

```
router.go      → route registration
handler.go     → request parsing, response writing
service.go     → business logic
repository.go  → DB queries
models.go      → domain structs
module.go      → dependency wiring
dto/           → request + response types
```

---

## pkg/ — Shared Utilities

### response

`pkg/response/json.go` — all handlers use these instead of calling `c.JSON` directly.

| Function | Status | Description |
|---|---|---|
| `Success(c, data)` | 200 | standard data response |
| `Created(c, data)` | 201 | resource created |
| `NoContent(c)` | 204 | no body |
| `Error(c, err)` | auto | maps `AppError` → correct status + safe message |
| `Paginated(c, data, meta)` | 200 | data + pagination envelope |

```go
// handler example
func (h *Handler) Get(c *gin.Context) {
    user, err := h.svc.GetUser(c.Request.Context(), id)
    if err != nil {
        response.Error(c, err)   // apperr.NotFound → 404, apperr.Internal → 500
        return
    }
    response.Success(c, user)
}
```

---

### validator

`pkg/validator/validator.go` — wraps `go-playground/validator/v10` with a singleton and human-readable error messages.

```go
type RegisterRequest struct {
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required,strong_password"`
    Name     string `json:"name"     validate:"required,min=2,max=100"`
}

// in handler
if errs := validator.Validate(&req); errs != nil {
    c.JSON(http.StatusBadRequest, gin.H{"errors": errs})
    return
}
```

Built-in rules with human messages: `required`, `email`, `min`, `max`, `uuid4`, `oneof`, `url`

Custom rule: `strong_password` — min 8 chars, at least one uppercase, lowercase, and digit.

Error output uses snake_case field names:
```json
{
  "errors": {
    "email": "must be a valid email address",
    "password": "must be at least 8 characters with uppercase, lowercase, and a number"
  }
}
```

---

### pagination

`pkg/pagination/pagination.go` — offset and cursor pagination helpers.

```go
// offset-based (most list endpoints)
page := pagination.ParsePage(c)     // reads ?page=1&page_size=20
items, total, _ := svc.List(ctx, page.Limit(), page.Offset())

response.Paginated(c, items, response.PaginationMeta{
    Page:       page.Page,
    PageSize:   page.PageSize,
    Total:      total,
    TotalPages: pagination.TotalPages(total, page.PageSize),
    HasNext:    page.Page < pagination.TotalPages(total, page.PageSize),
    HasPrev:    page.Page > 1,
})

// cursor-based (activity feeds, infinite scroll)
cursor := pagination.ParseCursor(c)  // reads ?after=<token>&page_size=20
```

Defaults: `page=1`, `page_size=20`, max `page_size=100`.

---

## Security Notes

- JWT secret must be overridden via `JWT_SECRET` env — default is intentionally weak
- Passwords hashed with Argon2id (64MB memory, 3 iterations) — `subtle.ConstantTimeCompare` for verify
- Refresh tokens stored as hashes in `sessions` table — raw token never persisted
- SSL defaults to `disable` for local dev — set `SSL_MODE=require` in production
- JWT signing uses HS256 — algorithm is validated on parse to prevent confusion attacks

---

## Refresh Token System

Refresh token lifecycle is handled by `internal/shared/core/session/`.

### Rotation — implemented

Every use of a refresh token issues a new one and revokes the old session row.

```
Client sends refresh token
  ↓
Hash token → DB lookup (GetByTokenHash)
  ↓
Check revoked_at IS NULL + expires_at > NOW()
  ↓
Revoke old session row
  ↓
Insert new session row with new token hash
  ↓
Return new access token + new refresh token
```

### Reuse detection — implemented

A revoked token being presented again triggers full account lockout.

```
Token hash found AND revoked_at IS NOT NULL
  ↓
slog.Warn("token reuse detected")
  ↓
RevokeAll(userID) — all devices signed out
  ↓
Return 401 — force full re-authentication
```

### Session revocation — implemented

| Scenario                   | Method                        |
|----------------------------|-------------------------------|
| User logs out              | `Revoke(sessionID)`           |
| Logout all devices         | `RevokeAll(userID)`           |
| Admin suspends user        | `RevokeAll(userID)`           |
| Token reuse detected       | `RevokeAll(userID)` automatic |
| Token expired              | `expires_at < NOW()` check    |

All logic uses the existing `sessions` schema — no new migrations needed.

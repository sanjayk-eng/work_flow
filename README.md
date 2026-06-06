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
- **Logging:** stdlib `slog` (JSON, structured, context-threaded)
- **Validation:** [go-playground/validator/v10](https://github.com/go-playground/validator)

---

## Project Structure

```
.
├── cmd/
│   └── api/
│       ├── main.go                               # Entry point — wires config, DB, logger, server
│       └── .env                                  # Local environment variables
│
├── internal/
│   ├── middleware/
│   │   ├── auth.go                               # Authenticate — JWT verify, sets user_id + email
│   │   ├── org.go                                # OrgContext — DB membership lookup, sets org_id + role
│   │   └── rbac.go                               # RequireRole — role-based access check
│   │
│   ├── modules/
│   │   ├── auth/
│   │   │   ├── dto/
│   │   │   │   ├── request.go                    # RegisterRequest (validated struct tags)
│   │   │   │   └── response.go                   # UserResponse, UserResponseInput, NewUserResponse()
│   │   │   ├── handler.go                        # Register — bind, validate, call service
│   │   │   ├── models.go                         # User, Profile domain structs
│   │   │   ├── module.go                         # New() wires repo→service→handler, Mount() registers routes
│   │   │   ├── repository.go                     # Repository interface (ExistsByEmail, CreateUser, WithTx…)
│   │   │   ├── repository_postgres.go            # Postgres implementation of Repository
│   │   │   ├── router.go                         # RegisterRoutes — POST /auth/register
│   │   │   └── service.go                        # Service interface + registration flow logic
│   │   ├── organization/                         # (placeholder)
│   │   └── user/                                 # (placeholder)
│   │
│   └── shared/
│       ├── apperr/
│       │   └── apperr.go                         # Typed domain errors — NotFound, Unauthorized, Conflict…
│       │
│       ├── core/
│       │   ├── config/
│       │   │   ├── config.go                     # Config, AppConfig, DBConfig, JWTConfig, LogConfig structs
│       │   │   ├── const.go                      # Env key constants (app, db, redis, jwt, log)
│       │   │   ├── env.go                        # godotenv loader
│       │   │   ├── loader.go                     # NewConfig() — reads env into structs
│       │   │   └── provider.go                   # Get() singleton via sync.Once
│       │   │
│       │   ├── database/
│       │   │   └── postgres/
│       │   │       ├── connections.go            # pgxpool setup, New(), Close()
│       │   │       ├── const.go                  # Pool defaults (max/min conns, timeouts)
│       │   │       ├── dsn.go                    # BuildDSN() — postgres connection string
│       │   │       ├── migration.go              # MigrateUp(), MigrateDown() via Goose
│       │   │       └── tx.go                     # WithTx(), WithTxSerializable() — transaction helpers
│       │   │
│       │   ├── logger/
│       │   │   └── logger.go                     # New(), WithContext(), FromContext() — slog JSON
│       │   │
│       │   ├── session/
│       │   │   ├── session.go                    # Session struct, Repository interface, GenerateRefreshToken()
│       │   │   └── service.go                    # Issue(), Rotate() + reuse detection, Revoke(), RevokeAll()
│       │   │
│       │   └── server/
│       │       └── server.go                     # Gin engine, Recovery middleware, logger injection, Run()
│       │
│       └── security/
│           ├── jwt/
│           │   ├── claims.go                     # Claims — user_id + email only
│           │   ├── config.go                     # JWT config struct (secret, expiry, issuer)
│           │   └── service.go                    # New(), CreateToken(), VerifyToken()
│           └── password/
│               ├── argon2.go                     # Pure generateHash() — argon2id
│               ├── config.go                     # Argon2 params (memory, iterations, parallelism)
│               └── password.go                   # Hasher — Hash(), Verify()
│
├── migrations/
│   ├── 20260605155734_tbl_user.sql               # users table + user_status enum
│   ├── 20260605155848_create_profiles.sql        # profiles table
│   ├── 20260605160022_create_email_verifications.sql  # email_verifications table
│   └── 20260605160339_create_sessions.sql        # sessions table — refresh token store
│
├── pkg/
│   ├── response/
│   │   └── json.go                               # Success(), Created(), Error(), Paginated(), PaginationMeta
│   ├── validator/
│   │   └── validator.go                          # Validate(), strong_password rule, snake_case error keys
│   └── pagination/
│       └── pagination.go                         # ParsePage(), ParseCursor(), TotalPages(), Page, Cursor
│
├── go.mod
├── go.sum
└── README.md
```

---

## Auth Architecture

### JWT — identity only

```
Claims { user_id, email, iss, iat, exp }
```

```go
svc := jwt.New([]byte(cfg.JWT.SecretKey))
token, _  := svc.CreateToken(userID, email)   // 15-min access token
claims, _ := svc.VerifyToken(tokenStr)         // returns Claims or sentinel error
```

Refresh tokens are random 32-byte values stored as SHA-256 hashes in the `sessions` table — never embedded in JWTs.

### 3-layer middleware chain

```
Request
  ↓
Authenticate(jwtSvc)           → verifies JWT → sets user_id, email in context
  ↓
OrgContext(orgRepo)             → reads X-Org-ID header → DB lookup → sets org_id, role
  ↓
RequireRole("admin", "member")  → checks role → 403 if not allowed
  ↓
Handler
```

### Context keys

| Key       | Set by         | Value              |
|-----------|----------------|--------------------|
| `user_id` | `Authenticate` | UUID string        |
| `email`   | `Authenticate` | user email         |
| `org_id`  | `OrgContext`   | UUID string        |
| `role`    | `OrgContext`   | e.g. admin, member |

---

## Registration Flow

`POST /api/v1/auth/register`

```
Handler (bind + validate)
  ↓
Service.Register()
  ├── 1. repo.ExistsByEmail()       → 409 Conflict if taken
  ├── 2. hasher.Hash(password)      → Argon2id hash
  ├── 3. repo.WithTx()              → begin transaction
  │       ├── repo.CreateUser()     → INSERT users
  │       └── repo.CreateProfile()  → INSERT profiles
  │   (rollback on any error)
  └── 4. dto.NewUserResponse()      → map User + Profile → response
```

Request:
```json
{
  "email":      "user@example.com",
  "password":   "SecurePass1",
  "first_name": "John",
  "last_name":  "Doe"
}
```

Response `201`:
```json
{
  "data": {
    "user": {
      "id":         "uuid",
      "email":      "user@example.com",
      "first_name": "John",
      "last_name":  "Doe",
      "status":     "pending",
      "created_at": "2026-06-06T..."
    }
  }
}
```

### Module layer dependency chain

```
module.go
  └── NewRepository(db)      → postgresRepository (implements Repository interface)
  └── NewService(repo, hasher) → service (implements Service interface)
  └── NewHandler(svc)        → Handler
  └── Mount(rg)              → RegisterRoutes
```

Service never holds `*postgres.DB` — it calls `repo.WithTx()` through the interface. DB is fully hidden behind the repository.

---

## Database Schema

| Table                  | Description                                          |
|------------------------|------------------------------------------------------|
| `users`                | Core user — email, password hash, status enum        |
| `profiles`             | User profile — first name, last name, avatar         |
| `email_verifications`  | Email verification tokens with expiry                |
| `sessions`             | Refresh token hashes — expires_at + revoked_at       |

---

## Refresh Token System

Handled by `internal/shared/core/session/`.

### Rotation
Every refresh issues a new token and revokes the old session row.

### Reuse detection
A revoked token presented again → `RevokeAll(userID)` → all devices signed out → 401.

### Revocation scenarios

| Scenario              | Method                        |
|-----------------------|-------------------------------|
| Logout                | `Revoke(sessionID)`           |
| Logout all devices    | `RevokeAll(userID)`           |
| Admin suspends user   | `RevokeAll(userID)`           |
| Reuse detected        | `RevokeAll(userID)` automatic |

---

## pkg/ — Shared Utilities

### response

```go
response.Success(c, data)          // 200
response.Created(c, data)          // 201
response.NoContent(c)              // 204
response.Error(c, err)             // apperr.AppError → correct status automatically
response.Paginated(c, data, meta)  // 200 with pagination envelope
```

### validator

```go
type RegisterRequest struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,strong_password"`
}

if errs := validator.Validate(&req); errs != nil {
    c.JSON(400, gin.H{"errors": errs})
}
// errs → map[string]string{"email": "must be a valid email address"}
```

Custom rule: `strong_password` — min 8 chars, uppercase + lowercase + digit.

### pagination

```go
page   := pagination.ParsePage(c)    // ?page=1&page_size=20
cursor := pagination.ParseCursor(c)  // ?after=<token>&page_size=20

items, total, _ := svc.List(ctx, page.Limit(), page.Offset())
response.Paginated(c, items, response.PaginationMeta{
    Page:       page.Page,
    PageSize:   page.PageSize,
    Total:      total,
    TotalPages: pagination.TotalPages(total, page.PageSize),
    HasNext:    page.Page < pagination.TotalPages(total, page.PageSize),
    HasPrev:    page.Page > 1,
})
```

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

## Security Notes

- JWT secret must be overridden via `JWT_SECRET` — default is intentionally weak
- Passwords hashed with Argon2id (64MB memory, 3 iterations) — `subtle.ConstantTimeCompare` on verify
- Refresh tokens stored as SHA-256 hashes — raw token never persisted
- JWT uses HS256 — signing method validated on parse to prevent algorithm confusion attacks
- SSL defaults to `disable` for local dev — set `SSL_MODE=require` in production

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
│       ├── main.go                                      # Entry point — wires config, DB, server, IAM module
│       └── .env                                         # Local environment variables
│
├── internal/
│   ├── middleware/
│   │   ├── auth.go                                      # Authenticate — JWT verify, sets user_id + email
│   │   ├── org.go                                       # OrgContext — DB membership lookup, sets org_id + role
│   │   └── rbac.go                                      # RequireRole — role-based access check
│   │
│   ├── modules/
│   │   ├── iam/                                         # Identity & Access Management domain
│   │   │   ├── module.go                                # Top-level IAM module — wires auth + user + session
│   │   │   │
│   │   │   ├── auth/                                    # Authentication sub-module
│   │   │   │   ├── dto/
│   │   │   │   │   ├── request.go                       # RegisterRequest, LoginRequest
│   │   │   │   │   └── response.go                      # UserResponse, LoginResponse, RefreshResponse
│   │   │   │   ├── handler.go                           # HTTP handlers — Register, Login, Refresh, EmailVerification
│   │   │   │   ├── models.go                            # User, EmailVerification domain structs
│   │   │   │   ├── module.go                            # NewModule(handler) + Mount()
│   │   │   │   ├── repository.go                        # Repository interface — users + email verifications
│   │   │   │   ├── repository_postgres.go               # Postgres implementation
│   │   │   │   ├── router.go                            # RegisterRoutes — /auth group
│   │   │   │   └── service.go                           # Service interface + full auth flow logic
│   │   │   │
│   │   │   ├── user/                                    # User profile sub-module
│   │   │   │   ├── dto/
│   │   │   │   │   ├── request.go                       # CreateProfileRequest
│   │   │   │   │   └── response.go                      # (reserved)
│   │   │   │   ├── models.go                            # Profile domain struct
│   │   │   │   ├── repository.go                        # Repository interface — profiles
│   │   │   │   ├── repository_postgres.go               # Postgres implementation
│   │   │   │   └── service.go                           # Service interface + CreateProfile logic
│   │   │   │
│   │   │   └── session/                                 # Session lifecycle sub-module
│   │   │       ├── models.go                            # Session domain struct
│   │   │       ├── repository.go                        # Repository interface — sessions CRUD
│   │   │       ├── repository_postgres.go               # Postgres implementation
│   │   │       └── service.go                           # Service interface — Create, GetByRefreshHash, Revoke, RevokeAll
│   │   │
│   │   └── organization/                                # (placeholder)
│   │
│   └── shared/
│       ├── apperr/
│       │   └── apperr.go                                # Typed domain errors — NotFound, Unauthorized, Conflict…
│       │
│       ├── core/
│       │   ├── config/
│       │   │   ├── config.go                            # Config, AppConfig, DBConfig, JWTConfig, LogConfig structs
│       │   │   ├── const.go                             # Env key constants
│       │   │   ├── env.go                               # godotenv loader
│       │   │   ├── loader.go                            # NewConfig() — reads env into structs
│       │   │   └── provider.go                          # Get() singleton via sync.Once
│       │   │
│       │   ├── database/
│       │   │   └── postgres/
│       │   │       ├── connections.go                   # pgxpool setup, New(), Close()
│       │   │       ├── const.go                         # Pool defaults
│       │   │       ├── dsn.go                           # BuildDSN()
│       │   │       ├── migration.go                     # MigrateUp(), MigrateDown() via Goose
│       │   │       └── tx.go                            # WithTx() transaction helper
│       │   │
│       │   └── server/
│       │       └── server.go                            # Gin engine, Recovery middleware, Run()
│       │
│       └── security/
│           ├── security.go                              # GenerateToken(), GenerateRefreshToken()
│           ├── jwt/
│           │   ├── claims.go                            # Claims — user_id + email only
│           │   ├── config.go                            # JWT config struct (secret, expiry, issuer)
│           │   └── service.go                           # New(), CreateToken(), VerifyToken()
│           └── password/
│               ├── argon2.go                            # Pure generateHash() — argon2id
│               ├── config.go                            # Argon2 params
│               └── password.go                          # Hasher — Hash(), Verify()
│
├── migrations/
│   ├── 20260605155734_tbl_user.sql                      # users table + user_status enum
│   ├── 20260605155848_create_profiles.sql               # profiles table
│   ├── 20260605160022_create_email_verifications.sql    # email_verifications table
│   └── 20260605160339_create_sessions.sql               # sessions table — refresh token store
│
├── pkg/
│   ├── response/
│   │   └── json.go                                      # Success(), Error(), Paginated()
│   ├── validator/
│   │   └── validator.go                                 # BindAndValidate(), strong_password rule
│   └── pagination/
│       └── pagination.go                                # ParsePage(), TotalPages()
│
├── go.mod
├── go.sum
└── README.md
```

---

## IAM Module Architecture

All identity-related logic is grouped under `internal/modules/iam/`. `main.go` only needs one call:

```go
iam.New(db, cfg.JWT).Mount(api)
```

### Sub-module responsibilities

| Sub-module  | Owns                                                   |
|-------------|--------------------------------------------------------|
| `auth`      | User registration, email verification, login, token refresh |
| `user`      | Profile creation and management                        |
| `session`   | Refresh token storage, rotation, and revocation        |

### Dependency graph

```
iam.Module
  ├── auth.Service  ──depends──▶  user.Service
  │                 ──depends──▶  session.Service
  │                 ──depends──▶  jwt.Service
  │                 ──depends──▶  password.Hasher
  ├── user.Service  ──depends──▶  user.Repository
  └── session.Service ─depends─▶  session.Repository
```

No circular dependencies — `user` and `session` know nothing about `auth`.

---

## Auth Architecture

### JWT — identity only

```
Claims { user_id, email, iss, iat, exp }
```

```go
svc := jwt.New([]byte(cfg.JWT.SecretKey))
token, _  := svc.CreateToken(userID, email, expiresAt)
claims, _ := svc.VerifyToken(tokenStr)
```

Refresh tokens are random 32-byte values stored as SHA-256 hashes in the `sessions` table — never embedded in JWTs.

### 3-layer middleware chain

```
Request
  ↓
Authenticate(jwtSvc)            → verifies JWT → sets user_id, email in context
  ↓
OrgContext(orgRepo)              → reads X-Org-ID header → DB lookup → sets org_id, role
  ↓
RequireRole("admin", "member")   → checks role → 403 if not allowed
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
auth.Service.Register()
  ├── 1. repo.ExistsByEmail()              → 409 Conflict if taken
  ├── 2. hasher.Hash(password)             → Argon2id hash
  ├── 3. db.WithTx()                       → begin transaction
  │       ├── repo.CreateUser()            → INSERT users
  │       ├── user.Service.CreateProfile() → INSERT profiles
  │       └── repo.CreateEmailVerification() → INSERT email_verifications
  │   (rollback on any error)
  └── 4. return UserResponse
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

Response `200`:
```json
{
  "data": {
    "id":         "uuid",
    "email":      "user@example.com",
    "created_at": "2026-06-07T..."
  }
}
```

---

## Session / Refresh Token System

Handled by `internal/modules/iam/session/`.

### Rotation
Every refresh call issues a new token and revokes the old session row.

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

## Database Schema

| Table                  | Description                                        |
|------------------------|----------------------------------------------------|
| `users`                | Core user — email, password hash, status enum      |
| `profiles`             | User profile — first name, last name, avatar       |
| `email_verifications`  | Email verification tokens with expiry              |
| `sessions`             | Refresh token hashes — expires_at + revoked_at     |

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

Migrations run automatically on startup.

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

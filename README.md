# work_flow

## Folder Structure

```text
work_flow/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── middleware/
│   │   └── auth.go
│   ├── moduler/
│   │   └── auth/
│   └── shared/
│       ├── config/
│       │   ├── config.go
│       │   ├── const.go
│       │   ├── env.go
│       │   ├── loader.go
│       │   └── provider.go
│       └── database/
│           └── postgres/
│               ├── connections.go
│               ├── const.go
│               ├── dsn.go
│               └── migration.go
├── migrations/
├── pkg/
├── go.mod
└── go.sum
```

# Golang REST API — Practice 3

A production-ready RESTful API built with Go, PostgreSQL, and Clean Architecture principles. Implements full CRUD for users with logging and API key authentication middleware.

---

## Architecture

The project follows **Clean Architecture** with three distinct layers:

```
Handler → Usecase → Repository
```

- **Handler layer** — receives HTTP requests, validates input, calls the Usecase layer, returns JSON responses with correct HTTP status codes.
- **Usecase layer** — contains business logic; calls the Repository layer through interfaces.
- **Repository layer** — performs all database operations using `sqlx`.

Layers communicate through **interfaces**, making each layer independently testable and replaceable.

---

## Folder Structure

```
golang/
├── cmd/
│   └── api/
│       └── main.go                        # Entry point — calls app.Run()
├── database/
│   └── migrations/
│       ├── 000001_init.down.sql           # Drops users table
│       └── 000002_init.up.sql             # Creates users table + seeds data
├── internal/
│   ├── app/
│   │   └── main.go                        # Wires all layers and starts the server
│   ├── handler/
│   │   ├── handler.go                     # Router setup + Healthcheck
│   │   └── users.go                       # User HTTP handlers (5 endpoints)
│   ├── middleware/
│   │   ├── auth.go                        # X-API-KEY validation (401 if invalid)
│   │   └── logging.go                     # Logs timestamp, method, endpoint, status
│   ├── usecase/
│   │   ├── usecase.go                     # UserUsecase interface + Usecases struct
│   │   └── users.go                       # UserUsecase implementation
│   └── repository/
│       ├── repository.go                  # UserRepository interface + Repositories struct
│       └── _postgres/
│           ├── postgres.go                # DB connection + AutoMigrate
│           └── users/
│               └── users.go              # All CRUD repository methods
├── pkg/
│   └── modules/
│       ├── configs.go                     # PostgreConfig struct
│       └── users.go                       # User struct + request types
├── go.mod
└── README.md
```

---

## User Model

The `User` struct contains **5 fields** (ID + Name + 3 extra):

| Field      | Type      | DB Column    |
|------------|-----------|--------------|
| ID         | int       | id           |
| Name       | string    | name         |
| Email      | string    | email        |
| Age        | int       | age          |
| CreatedAt  | time.Time | created_at   |

---

## Prerequisites

- Go 1.21+
- PostgreSQL running locally on port `5432`
- Database `mydb` created
- User `postgres` with password `postgres`

```bash
# Create the database (if not exists)
psql -U postgres -c "CREATE DATABASE mydb;"
```

---

## How to Run

```bash
# From the project root
go mod tidy
go run cmd/api/main.go
```

The server starts on **`:8080`** and auto-runs database migrations on startup.

---

## API Endpoints

All endpoints (except `/health`) require the header: `X-API-KEY: mysecretapikey`

All responses are `Content-Type: application/json`.

### Healthcheck

```bash
GET /health
curl http://localhost:8080/health
```

Response:
```json
{"status": "ok"}
```

---

### Get All Users

```bash
curl -H "X-API-KEY: mysecretapikey" http://localhost:8080/users
```

Response `200 OK`:
```json
[
  {"id": 1, "name": "John Doe", "email": "john@example.com", "age": 30, "created_at": "..."}
]
```

---

### Get User by ID

```bash
curl -H "X-API-KEY: mysecretapikey" http://localhost:8080/users/1
```

Response `200 OK`:
```json
{"id": 1, "name": "John Doe", "email": "john@example.com", "age": 30, "created_at": "..."}
```

Response `404 Not Found` (if ID doesn't exist):
```json
{"error": "GetUserByID: user with id=99 not found: sql: no rows in result set"}
```

---

### Create User

```bash
curl -X POST http://localhost:8080/users \
  -H "X-API-KEY: mysecretapikey" \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Smith", "email": "jane@example.com", "age": 25}'
```

Response `201 Created`:
```json
{"id": 2, "message": "user created successfully"}
```

---

### Update User

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "X-API-KEY: mysecretapikey" \
  -H "Content-Type: application/json" \
  -d '{"name": "John Updated", "email": "updated@example.com", "age": 31}'
```

Response `200 OK`:
```json
{"message": "user updated successfully"}
```

Response `404 Not Found` (0 rows affected):
```json
{"error": "UpdateUser: user with id=99 does not exist"}
```

---

### Delete User

```bash
curl -X DELETE http://localhost:8080/users/1 \
  -H "X-API-KEY: mysecretapikey"
```

Response `200 OK`:
```json
{"message": "user deleted successfully", "rows_affected": 1}
```

---

## Middleware

### Logger (`internal/middleware/logging.go`)

Wraps every request. After the handler responds, logs to stdout using the standard `log` package (not `fmt`):

```
[2026-02-19T10:00:00Z] method=GET endpoint=/users status=200 duration=1.2ms
```

Required fields: `timestamp`, `http method`, `endpoint name`, `status code`.

### Auth (`internal/middleware/auth.go`)

Checks the `X-API-KEY` header on every request (except `/health`).

- **Missing or invalid key** → `401 Unauthorized` with JSON error body.
- **Valid key** (`mysecretapikey`) → request proceeds to the handler.

```bash
# This returns 401
curl http://localhost:8080/users

# This succeeds
curl -H "X-API-KEY: mysecretapikey" http://localhost:8080/users
```

---

## Error Handling

| Scenario                         | Status Code |
|----------------------------------|-------------|
| Missing/invalid API key          | 401         |
| Invalid ID in path               | 400         |
| Invalid JSON body                | 400         |
| Required field missing           | 400         |
| User not found (GET/PUT/DELETE)  | 404         |
| Database / internal error        | 500         |

---

## Definition of Done Checklist

- [x] `go run cmd/api/main.go` starts the server on `:8080`
- [x] All endpoints return JSON with `Content-Type: application/json`
- [x] Correct HTTP status codes used everywhere
- [x] UserRepository has 5 functions: GetUsers, GetUserByID, CreateUser, UpdateUser, DeleteUser
- [x] Handler → Usecase → Repository layer chain via interfaces
- [x] Healthcheck endpoint at `GET /health`
- [x] Auth middleware blocks requests without valid `X-API-KEY` (401)
- [x] Logger middleware logs every request (timestamp, method, endpoint)
- [x] Standard `log` package used (not `fmt`) for logging
- [x] User struct has 5 fields (ID, Name + 3 extra: Email, Age, CreatedAt)
- [x] Migration files in `database/migrations/`
- [x] Insert returns newly generated ID
- [x] Update checks RowsAffected, returns error if 0 rows
- [x] Delete checks RowsAffected, returns error if 0 rows
- [x] GetUserByID returns nil + informative error if ID not found

---

## Demo Video Script

See `DEMO_SCRIPT.md` for the full demo walkthrough.

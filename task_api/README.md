# Task API (Go, net/http)

Minimal backend service for managing tasks (todos/issues) using Go standard library only.

## Features

- REST API on Go (`net/http`)
- In-memory storage (`map[int]Task`)
- CRUD operations for tasks
- API Key authentication middleware
- Logging middleware
- JSON-only responses
- Filtering tasks by status (`done=true`)
- Concurrency safe (`sync.Mutex`)
- Proper HTTP status codes

## Tech Stack

- Language: Go
- Frameworks: ❌ None (standard library only)
- Storage: In-memory
- Port: 8080

## Project Structure
task-api/
│
├── cmd/api/main.go # Entry point
├── internal/
│ ├── handlers/ # HTTP handlers
│ ├── middleware/ # API Key & Logging middleware
│ ├── models/ # Data models
│ └── storage/ # In-memory storage
│
├── go.mod
└── README.md

## ▶ How to Run
```bash
go mod tidy
go run ./cmd/api
```
Server starts on:
http://localhost:8080

## All requests require API key:

- Header: X-API-KEY
- Value: secret12345

If missing or invalid → 401 Unauthorized

## API Endpoints

Get all tasks
- GET /tasks

Get task by ID
- GET /tasks?id=1

Filter tasks
- GET /tasks?done=true

Create task
- POST /tasks

  Body:
```
{
"title": "Write unit tests"
}
```

Update task status
- PATCH /tasks?id=1

Body:
```{
"done": true
}
```
Delete task

- DELETE /tasks?id=1

🧪 Example Response
```
{
"id": 1,
"title": "Write unit tests",
"done": false
}
```
📝 Logging Example
```
2026-01-29T14:49:05 POST /tasks request received
```
 Definition of Done

- JSON-only responses
- Correct HTTP status codes
- Middleware for auth & logging
- Thread-safe in-memory storage
- Clean project structure
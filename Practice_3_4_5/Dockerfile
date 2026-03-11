FROM golang:1.25.5-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o myapp ./cmd/api/main.go

FROM alpine
COPY --from=builder app/myapp /myapp
COPY --from=builder /app/database ./database
ENTRYPOINT ["/myapp"]
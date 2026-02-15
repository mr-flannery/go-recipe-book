FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o recipe-book ./src/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/recipe-book .
COPY src/templates ./src/templates
COPY src/static ./src/static
COPY src/db/migrations ./src/db/migrations

ENV APP_BASE_PATH=/app
ENV PORT=8080

CMD ["./recipe-book"]

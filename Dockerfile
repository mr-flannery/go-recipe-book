FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG COMMIT_HASH=unknown
RUN CGO_ENABLED=0 go build -ldflags="-X github.com/mr-flannery/go-recipe-book/src/handlers.CommitHash=${COMMIT_HASH}" -o recipe-book ./src/main.go

FROM alpine:latest
WORKDIR /app

# Install ffmpeg and python/pip for yt-dlp (needed for video audio extraction)
RUN apk add --no-cache ffmpeg python3 py3-pip && \
    pip3 install --no-cache-dir --break-system-packages yt-dlp

COPY --from=builder /app/recipe-book .
COPY src/templates ./src/templates
COPY src/static ./src/static
COPY src/db/migrations ./src/db/migrations

ENV APP_BASE_PATH=/app
ENV PORT=8080

CMD ["./recipe-book"]

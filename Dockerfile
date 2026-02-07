FROM golang:1.21-alpine as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o recipe-book ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/recipe-book .
COPY src/templates ./src/templates
COPY src/static ./src/static
COPY src/db ./src/db
ENV PORT=8080
CMD ["./recipe-book"]

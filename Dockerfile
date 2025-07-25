FROM golang:1.21-alpine as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o recipe-book ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/recipe-book .
COPY templates ./templates
COPY static ./static
COPY db ./db
ENV PORT=8080
CMD ["./recipe-book"]

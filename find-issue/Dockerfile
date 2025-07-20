# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY main.go ./
RUN go build -o find-issue main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /

COPY --from=builder /app/find-issue /find-issue

ENTRYPOINT ["/find-issue"]

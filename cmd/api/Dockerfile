FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./resources/ ./resources/

RUN go mod download
RUN go mod tidy

RUN GOOS=linux go build -o api ./cmd/api/

FROM debian:bookworm-slim
WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/resources/ ./resources
CMD ["./api"]
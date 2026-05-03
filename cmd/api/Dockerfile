FROM golang:tip-alpine3.22

WORKDIR /app

COPY ../go.mod ./

RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./resources/ ./resources/

RUN GOOS=linux go build -o api ./cmd/api/

CMD [ "./api" ]
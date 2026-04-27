FROM golang:tip-alpine3.22

WORKDIR /app

COPY ../go.mod ./

RUN go mod download

COPY ./cmd/api/*.go ./ 

RUN ls

RUN GOOS=linux go build -o api .

CMD [ "./api" ]
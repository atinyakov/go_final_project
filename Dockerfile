FROM golang:1.21 AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /my_app ./cmd/app

FROM ubuntu:latest

WORKDIR /root/
COPY --from=builder /my_app .
COPY ./web ./web 
COPY .env . 

CMD ["./my_app"]

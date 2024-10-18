FROM golang:1.21

WORKDIR /usr/src/app 

COPY go.mod go.sum ./

RUN go mod download

COPY *.go *.db ./
COPY ./controllers ./controllers
COPY ./services ./services
COPY ./nextdate ./nextdate
COPY ./models ./models
COPY ./web ./web
COPY .env .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /my_app

CMD ["/my_app"]

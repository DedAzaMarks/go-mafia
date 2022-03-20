FROM golang:latest

RUN apt-get update
WORKDIR /go/src/server/mafia-rpc
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
EXPOSE 42069

CMD ["go", "run", "/server/main.go"]

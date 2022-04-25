FROM golang:alpine

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -buildvcs=false -v -o /usr/local/bin/ ./...

CMD ["/usr/local/bin/go-api"]

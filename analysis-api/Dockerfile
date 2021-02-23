FROM golang:1.16

WORKDIR src/api
COPY . .

RUN go mod tidy
RUN go vet cmd/api.go
RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT [ "cmd"  ]
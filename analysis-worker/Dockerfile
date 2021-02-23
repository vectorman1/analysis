FROM golang:1.16

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go mod tidy
RUN go vet

RUN go build -o worker .

CMD [ "./worker" ]

EXPOSE 6969:6969
FROM golang:1.25.6

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /server ./cmd/server

EXPOSE 8080

ENTRYPOINT ["/server"]

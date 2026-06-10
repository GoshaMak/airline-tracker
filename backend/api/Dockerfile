FROM golang:1.25.6

WORKDIR /backend

COPY . .

WORKDIR /backend/api

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /bin/api ./cmd/api

EXPOSE 8080

ENTRYPOINT ["/bin/api"]

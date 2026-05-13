FROM golang:1.25.6

WORKDIR /shared

COPY shared/go.mod .
RUN go mod download

COPY shared .

WORKDIR /api

COPY services/api/go.mod services/api/go.sum ./
RUN go mod download

COPY services/api .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /bin/api ./cmd/server

EXPOSE 8080

ENTRYPOINT ["/bin/api"]

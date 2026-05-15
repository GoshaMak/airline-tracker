FROM golang:1.25.6

WORKDIR /shared

COPY shared/go.mod .
RUN go mod download

COPY shared .

WORKDIR /notifier

COPY services/notifier/go.mod services/notifier/go.sum ./
RUN go mod download

COPY services/notifier .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /bin/notifier ./cmd/notifier

ENTRYPOINT ["/bin/notifier"]

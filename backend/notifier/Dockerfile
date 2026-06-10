FROM golang:1.25.6

WORKDIR /backend

COPY . .

WORKDIR /backend/notifier

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /bin/notifier ./cmd/notifier

ENTRYPOINT ["/bin/notifier"]

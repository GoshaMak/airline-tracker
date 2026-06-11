FROM golang:1.26 as builder

WORKDIR /backend

COPY go.work go.work.sum ./
COPY shared/go.mod ./shared/
COPY api/go.mod api/go.sum ./api/
COPY notifier/go.mod notifier/go.sum ./notifier/

RUN --mount=type=cache,target=/go/pkg/mod \
    go work sync

COPY . .

WORKDIR /backend/notifier

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux \
    go build -o /bin/notifier ./cmd/notifier


FROM alpine:latest

COPY --from=builder /bin/notifier /bin/notifier

ENTRYPOINT ["/bin/notifier"]

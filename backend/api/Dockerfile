FROM golang:1.26 as builder

WORKDIR /backend

COPY go.work go.work.sum ./
COPY shared/go.mod ./shared/
COPY notifier/go.mod notifier/go.sum ./notifier/
COPY api/go.mod api/go.sum ./api/

RUN --mount=type=cache,target=/go/pkg/mod \
    go work sync

COPY . .

WORKDIR /backend/api

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux \
    go build -o /bin/api ./cmd/api


FROM alpine:latest

COPY --from=builder /bin/api /bin/api

EXPOSE 8080

ENTRYPOINT ["/bin/api"]

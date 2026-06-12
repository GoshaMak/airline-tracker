FROM golang:1.26 AS builder

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
    go build -o /bin/migrate ./cmd/migrate

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux \
    go build -o /bin/api ./cmd/api


FROM alpine:latest

RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add supervisor

COPY api/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

COPY --from=builder /bin/api /bin/api
COPY --from=builder /bin/migrate /bin/migrate

EXPOSE 8080

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]

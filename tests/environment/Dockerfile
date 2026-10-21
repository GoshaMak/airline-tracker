FROM golang:1.26.5 AS builder
WORKDIR /backend
COPY backend/ ./

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /bin/api ./api/cmd/api && \
    CGO_ENABLED=0 go build -o /bin/create-admin ./api/cmd/create-admin && \
    CGO_ENABLED=0 go build -o /bin/publisher ./api/cmd/publisher && \
    CGO_ENABLED=0 go build -o /bin/api-migrate ./api/cmd/migrate && \
    CGO_ENABLED=0 go build -o /bin/notifier ./notifier/cmd/notifier && \
    CGO_ENABLED=0 go build -o /bin/notifier-migrate ./notifier/cmd/migrate

COPY tests/environment/go.mod tests/environment/go.sum /test-support/
COPY tests/environment/cmd /test-support/cmd
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -C /test-support -mod=readonly -o /bin/drain ./cmd/kafka-drain

FROM builder AS integration-build
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go test -c -tags=integration -o /bin/integration-tests ./api/tests/integration

FROM alpine:3.23.3 AS integration-tests
COPY --from=integration-build /bin/integration-tests /bin/integration-tests
CMD ["/bin/integration-tests", "-test.v", "-test.count=1", "-test.timeout=60s"]

FROM builder AS e2e-build
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go test -c -tags=e2e -o /bin/e2e-tests ./api/tests/e2e

FROM alpine:3.23.3 AS e2e-tests
COPY --from=e2e-build /bin/e2e-tests /bin/e2e-tests
CMD ["/bin/e2e-tests", "-test.v", "-test.count=1", "-test.timeout=90s"]

FROM alpine:3.23.3

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /bin/api /bin/publisher /bin/api-migrate /bin/notifier /bin/notifier-migrate /bin/
COPY --from=builder /bin/drain /bin/drain
COPY --from=builder /bin/create-admin /bin/create-admin

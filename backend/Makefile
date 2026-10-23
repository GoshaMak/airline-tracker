API_PREFIX := ./api
NOTIFIER_PREFIX := ./notifier
PACKAGE ?= ./internal/flight/usecase
COMPOSE_ARGS := -f docker-compose.yml
DEBUG_COMPOSE_ARGS := -f infra-compose.debug.yml

.PHONY: up
up:
	docker compose $(COMPOSE_ARGS) up --build -d

.PHONY: replication-check
replication-check:
	docker compose $(COMPOSE_ARGS) run --rm --no-deps --entrypoint /bin/sh replication-setup /replication/verify.sh

.PHONY: create-admin
create-admin:
	docker compose $(COMPOSE_ARGS) run --build --rm --no-deps -e ADMIN_EMAIL -e ADMIN_PASSWORD api /bin/create-admin

.PHONY: test
test:
	sh scripts/run-allure.sh

export ALLURE_RESULTS_DIR ?= $(CURDIR)/allure-results
export ALLURE_LABEL_EPIC ?= unit

.PHONY: test-unit
test-unit:
	@cd $(API_PREFIX) && sh ../scripts/run-filtered-go-test.sh go test -v -count=1 -coverprofile=coverage.out ./...
	cd $(API_PREFIX) && python3 ../scripts/render-coverage.py
	@cd $(NOTIFIER_PREFIX) && sh ../scripts/run-filtered-go-test.sh go test -v -count=1 ./...
	@cd shared && sh ../scripts/run-filtered-go-test.sh go test -v -count=1 ./...

.PHONY: test-name
test-name:
	@cd $(API_PREFIX) && go test ./... -count=1 -run '^$(TEST)$$'


.PHONY: gobco
gobco:
	sh scripts/run-gobco.sh "$(PACKAGE)"

.PHONY: format
format:
	goimports -w $(API_PREFIX)/
	goimports -w $(NOTIFIER_PREFIX)/
	cd $(API_PREFIX) && go mod tidy
	cd $(NOTIFIER_PREFIX) && go mod tidy

.PHONY: debug
debug: format
debug: COMPOSE_ARGS += $(DEBUG_COMPOSE_ARGS)
debug: up

.PHONY: release
release: up

.PHONY: down
down:
	docker compose down -v

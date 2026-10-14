API_PREFIX := ./api
NOTIFIER_PREFIX := ./notifier
PACKAGE ?= ./internal/flight/usecase

HANDLERS := $(shell find $(API_PREFIX)/internal -type d -name handler | paste -sd ",")
DTOS := $(shell find $(API_PREFIX)/internal -type d -name dto | paste -sd ",")

.PHONY: swagger
swagger:
	swag init -g server.go -o $(API_PREFIX)/cmd/docs --parseInternal=true \
		--dir $(API_PREFIX)/internal/server/,$(HANDLERS),$(DTOS)

.PHONY: up
up:
	docker compose up --build -d

.PHONY: test
test:
	sh scripts/run-allure.sh

.PHONY: test-name
test-name:
	@cd $(API_PREFIX) && go test ./... -count=1 -run '^$(TEST)$$' 2>&1 | grep -v '\[no test files\]' | grep -v 'coverage: 0.0% of statements' | grep -v '\[no tests to run\]'


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
debug: format swagger up

.PHONY: release
release: up

.PHONY: down
down:
	docker compose down -v

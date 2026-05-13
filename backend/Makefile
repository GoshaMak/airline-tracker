API_PREFIX := services/api

HANDLERS := $(shell find $(API_PREFIX)/internal -type d -name handler | paste -sd ",")
DTOS := $(shell find $(API_PREFIX)/internal -type d -name dto | paste -sd ",")

.PHONY : swagger
swagger :
	swag init -g server.go -o $(API_PREFIX)/cmd/docs --parseInternal=true \
		--dir $(API_PREFIX)/internal/server/,$(HANDLERS),$(DTOS)

.PHONY : up
up: format
up: swagger
up:
	docker compose up --build -d

.PHONY : down
down:
	docker compose down -v

.PHONY : test
test:
	GO111MODULE=on go test -cover -count=1 ./$(API_PREFIX)/... 2>&1 | grep -v '\[no test files\]'

.PHONY : format
format:
	goimports -w $(API_PREFIX)/

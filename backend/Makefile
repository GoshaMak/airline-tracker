HANDLERS := $(shell find internal -type d -name handler | paste -sd ",")
DTOS := $(shell find internal -type d -name dto | paste -sd ",")

.PHONY : swagger
swagger :
	swag init -g server.go -o cmd/docs --parseInternal=true \
		--dir internal/server/,$(HANDLERS),$(DTOS)

.PHONY : up
up: format
up: swagger
up:
	docker compose up --build -d --force-recreate

.PHONY : down
down:
	docker compose down -v

.PHONY : test
test:
	GO111MODULE=on go test -cover -count=1 ./... 2>&1 | grep -v '\[no test files\]'

.PHONY : format
format:
	goimports -w .

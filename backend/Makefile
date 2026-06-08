API_PREFIX := services/api
NOTIFIER_PREFIX := services/notifier

HANDLERS := $(shell find $(API_PREFIX)/internal -type d -name handler | paste -sd ",")
DTOS := $(shell find $(API_PREFIX)/internal -type d -name dto | paste -sd ",")

.PHONY : swagger
swagger :
	swag init -g server.go -o $(API_PREFIX)/cmd/docs --parseInternal=true \
		--dir $(API_PREFIX)/internal/server/,$(HANDLERS),$(DTOS)

.PHONY : up
up :
	docker compose up --build -d

.PHONY : test
test :
	GO111MODULE=on go test -cover -count=1 ./$(API_PREFIX)/... 2>&1 | grep -v '\[no test files\]'

.PHONY : format
format :
	goimports -w $(API_PREFIX)/
	goimports -w $(NOTIFIER_PREFIX)/
	cd $(API_PREFIX) && go mod tidy
	cd $(NOTIFIER_PREFIX) && go mod tidy

.PHONY : debug
debug : format swagger up

release : up

.PHONY : down
down :
	docker compose down -v

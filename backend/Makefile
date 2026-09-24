API_PREFIX := ./api
NOTIFIER_PREFIX := ./notifier
ALLURE_VERSION := 2.46.1
ALLURE_ARCHIVE := .tools/allure-$(ALLURE_VERSION).tgz
ALLURE_BIN := $(abspath .tools/allure-$(ALLURE_VERSION)/bin/allure)
ALLURE_SHA256 := 37428164f310c8d137cfdd2c3a25d049b233261ddab13e1bd4dfd7627b5cbf8d

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
	GO111MODULE=on go test -cover -count=1 ./$(NOTIFIER_PREFIX)/... 2>&1 | grep -v '\[no test files\]'

.PHONY : test-offline
test-offline :
	cd $(API_PREFIX) && GOCACHE=/tmp/airline-tracker-go-build GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local ALLURE_RESULTS_DIR=/tmp/airline-tracker-allure-test-offline go test -count=1 ./...
	cd $(NOTIFIER_PREFIX) && GOCACHE=/tmp/airline-tracker-go-build GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local ALLURE_RESULTS_DIR=/tmp/airline-tracker-allure-test-offline go test -count=1 ./...

.PHONY : allure-install
allure-install : $(ALLURE_BIN)

$(ALLURE_BIN) :
	mkdir -p .tools
	curl -fL --retry 2 https://github.com/allure-framework/allure2/releases/download/$(ALLURE_VERSION)/allure-$(ALLURE_VERSION).tgz -o $(ALLURE_ARCHIVE)
	printf '%s  %s\n' '$(ALLURE_SHA256)' '$(ALLURE_ARCHIVE)' | sha256sum -c -
	tar -xzf $(ALLURE_ARCHIVE) -C .tools
	$(ALLURE_BIN) --version

.PHONY : allure-results
allure-results :
	sh scripts/run-allure.sh results

.PHONY : allure-deps
allure-deps :
	cd $(API_PREFIX) && GOWORK=off GOCACHE=/tmp/airline-tracker-go-build go mod download
	cd $(NOTIFIER_PREFIX) && GOWORK=off GOCACHE=/tmp/airline-tracker-go-build go mod download

.PHONY : allure-report
allure-report : allure-install allure-deps
	ALLURE_BIN=$(ALLURE_BIN) sh scripts/run-allure.sh report

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

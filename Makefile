CONTROLLERS := $(shell find internal -type d -name controller | paste -sd ",")
DTOS := $(shell find internal -type d -name dto | paste -sd ",")

.PHONY : swagger debug

swagger :
	swag init -g app.go -o cmd/docs --parseInternal=true \
		--dir internal/app/,$(CONTROLLERS),$(DTOS)

debug : swagger
debug : cmd/server/main.go
	go run $<

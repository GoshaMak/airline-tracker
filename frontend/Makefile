COMPOSE_ARGS := -f docker-compose.yml
DEBUG_COMPOSE_ARGS := -f docker-compose.debug.yml

.PHONY: up
up:
	docker compose $(COMPOSE_ARGS) up --build -d

.PHONY: debug
debug: COMPOSE_ARGS += $(DEBUG_COMPOSE_ARGS)
debug: up

.PHONY: release
release: up

.PHONY: down
down:
	docker compose $(COMPOSE_ARGS) down

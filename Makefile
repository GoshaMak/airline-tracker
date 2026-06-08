BACKEND_DIR := backend

.PHONY: backend-%
backend-%:
	$(MAKE) -C $(BACKEND_DIR) $*

.PHONY: frontend-up
frontend-up:
	docker compose -f frontend/docker-compose.yml up --build -d

.PHONY: frontend-down
frontend-down:
	docker compose -f frontend/docker-compose.yml down

.PHONY: fill-db
fill-db:
	docker compose -p api --env-file utils/.env -f utils/docker-compose.yml up --build -d

.PHONY: debug
debug: backend-debug
debug: fill-db
debug: frontend-up

.PHONY: release
release: backend-release
release: frontend-up

.PHONY: down
down: frontend-down
down: backend-down
down:
	docker compose -p api -f utils/docker-compose.yml down -v

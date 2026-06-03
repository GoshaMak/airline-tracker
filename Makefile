.PHONY: backend-up
backend-up:
	$(MAKE) -C backend up

.PHONY: backend-down
backend-down:
	$(MAKE) -C backend down

.PHONY: frontend-up
frontend-up:
	docker compose -f frontend/docker-compose.yml up --build -d

.PHONY: frontend-down
frontend-down:
	docker compose -f frontend/docker-compose.yml down

.PHONY: fill-db
fill-db:
	docker compose -p api --env-file utils/.env -f utils/docker-compose.yml up --build -d

.PHONY: all-up
all-up: backend-up
all-up: fill-db
all-up: frontend-up

.PHONY: all-down
all-down: frontend-down
all-down: backend-down
all-down:
	docker compose -p api -f utils/docker-compose.yml down -v

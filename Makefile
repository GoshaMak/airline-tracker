BACKEND_DIR := backend
FRONTEND_DIR := frontend

.PHONY: test
test:
	$(MAKE) -C $(BACKEND_DIR) test

.PHONY: load-test
load-test:
	k6 run --no-usage-report --new-machine-readable-summary=false tests/load/api.js

.PHONY: backend-%
backend-%:
	$(MAKE) -C $(BACKEND_DIR) $*

.PHONY: frontend-%
frontend-%:
	$(MAKE) -C $(FRONTEND_DIR) $*

.PHONY: fill-db
fill-db:
	docker compose -p api --env-file utils/.env -f utils/docker-compose.yml up --build -d

.PHONY: debug
debug: backend-debug
debug: fill-db
debug: frontend-debug

.PHONY: release
release: backend-release
release: frontend-release

.PHONY: down
down: frontend-down
down: backend-down
down:
	docker compose -p api -f utils/docker-compose.yml down -v

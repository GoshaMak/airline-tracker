BACKEND_DIR := backend
FRONTEND_DIR := frontend
RUN_ID ?= local
TEST_ENV = . tests/environment/stand.sh; environment_init "$(RUN_ID)";

.PHONY: test-env-up test-env-check test-env-reset test-env-down
test-env-up:
	$(TEST_ENV) environment_up

test-env-check:
	$(TEST_ENV) environment_check

test-env-reset:
	$(TEST_ENV) environment_reset

test-env-down:
	$(TEST_ENV) environment_down

.PHONY: test-integration
test-integration:
	$(TEST_ENV) environment_integration

.PHONY: test-e2e
test-e2e:
	$(TEST_ENV) environment_e2e

.PHONY: test-ci
test-ci:
	. tests/ci/run.sh; ci_init "$(RUN_ID)"; ci_all

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

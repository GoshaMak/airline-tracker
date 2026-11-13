BACKEND_DIR := backend
FRONTEND_DIR := frontend
RUN_ID ?= local
TEST_ENV = . tests/environment/stand.sh; environment_init "$(RUN_ID)";

.PHONY: test-env-up
test-env-up:
	$(TEST_ENV) environment_up

.PHONY: test-env-check
test-env-check:
	$(TEST_ENV) environment_check

.PHONY: test-env-reset
test-env-reset:
	$(TEST_ENV) environment_reset

.PHONY: test-env-down
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

.PHONY: test-traffic
test-traffic:
	$(TEST_ENV) environment_traffic

.PHONY: test
test:
	$(MAKE) -C $(BACKEND_DIR) test

.PHONY: load-test
load-test:
	k6 run --no-usage-report --new-machine-readable-summary=false tests/load/api.js

LOAD_ENV = export TEST_COMPOSE_EXTRA=tests/load/compose.yml MONITORING_CPUS=$${MONITORING_CPUS:-1} GRAFANA_ROOT_URL=$${GRAFANA_ROOT_URL:-http://localhost:$${LOAD_UI_PORT:-3001}/monitoring/}; $(TEST_ENV)
SCENARIO ?= ramp

.PHONY: load-up
load-up:
	$(LOAD_ENV) \
	environment_up; \
	dc exec -T -e E2E_ADMIN_PASSWORD="$$E2E_ADMIN_PASSWORD" postgres psql -X -v ON_ERROR_STOP=1 -U postgres -d airline_tracker < tests/load/seed.sql; \
	dc up -d --wait --wait-timeout 120 prometheus grafana loki tempo alloy load-ui cadvisor kafka-exporter postgres-exporter

.PHONY: load-run
load-run:
	$(LOAD_ENV) \
	mkdir -p "$$results/load"; \
	date +%s > "$$results/load/start"; \
	status=0; \
	dc run --rm --user "$$(id -u):$$(id -g)" k6 run --no-usage-report \
	    -o experimental-prometheus-rw --tag run_id=$(RUN_ID) --tag profile=$(SCENARIO) \
	    /scripts/$(SCENARIO).js || status=$$?; \
	date +%s > "$$results/load/end"; \
	exit "$$status"

.PHONY: load-export
load-export:
	$(LOAD_ENV) dc run --rm --no-deps --user "$$(id -u):$$(id -g)" tools python /load/export.py

.PHONY: load-down
load-down:
	$(LOAD_ENV) environment_down

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

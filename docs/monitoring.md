# Наблюдаемость в Docker

Grafana доступна по `http://localhost:8080/monitoring/`. Источники Prometheus/Loki.

## Состав

Prometheus опрашивает сервисы по Docker DNS каждые 15 секунд. Go-метрики доступны на внутреннем `:9090/metrics`, nginx stub_status -- на внутреннем `:8081/status`.

Grafana Alloy обнаруживает контейнеры обоих Compose-проектов по метке `airline_tracker.logs=true` и читает stdout/stderr через Docker API. В Loki добавляются `app="airline_tracker"`, `service`, `container`, `project`. Метки не зависят от ID контейнера.

## Запуск

Локальный вход по умолчанию -- `admin` / `admin`. Параметры читаются из `backend/.env`. `GRAFANA_ROOT_URL` меняется при публикации на другом домене/порту; по умолчанию используется адрес существующего ingress localhost:8080.

## Доступ к Docker API

Для логов Alloy использует socket Docker Engine `/var/run/docker.sock`.

Alloy запускается от root для чтения socket.

## Проверка конфигов

Проверка конфигов без запуска серверов мониторинга:

```bash
docker compose -f backend/docker-compose.yml run --rm --no-deps --entrypoint promtool prometheus check config /etc/prometheus/prometheus.yml
docker compose -f backend/docker-compose.yml run --rm --no-deps alloy validate /etc/alloy/alloy.cfg
docker compose -f backend/docker-compose.yml run --rm --no-deps loki -config.file=/etc/loki/config.yml -verify-config=true
```

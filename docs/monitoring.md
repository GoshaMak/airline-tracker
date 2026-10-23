# Наблюдаемость в Docker

Grafana доступна по `http://localhost:8080/monitoring/`. Источники: Prometheus (метрики), Loki (логи) и Tempo (распределённые трассировки).

## Состав

Prometheus опрашивает сервисы по Docker DNS каждые 15 секунд. Go-метрики доступны на внутреннем `:9090/metrics`, nginx stub_status -- на внутреннем `:8081/status`.

Grafana Alloy обнаруживает контейнеры обоих Compose-проектов по метке `airline_tracker.logs=true` и читает stdout/stderr через Docker API. В Loki добавляются `app="airline_tracker"`, `service`, `container`, `project`. Для Go-сервисов метка `app_service` имеет значение `api`, `publisher` или `notifier` и связывает логи с ресурсом `service.name` в трассировке.

API, publisher и notifier отправляют spans по OTLP/HTTP в Alloy (`http://alloy:4318`), затем Alloy передаёт их в Tempo. Порты OTLP и Tempo остаются внутри сети Compose. Трассировки сохраняются в Docker volume `tempo_data`, срок хранения — семь дней. Логи по-прежнему идут через stdout → Alloy → Loki.

## Как найти запрос и связанные операции

API принимает заголовок `X-Request-ID` длиной до 128 символов: латинские буквы, цифры, точка, дефис и подчёркивание. Для отсутствующего или некорректного значения создаётся UUID. В ответе API, включая HTTP 503 при истечении бюджета обработки, возвращается `X-Request-ID`.

Входящие `traceparent` и `tracestate` обрабатываются как W3C trace context. Контекст сохраняется вместе с outbox и уведомлением и передаётся в Kafka headers, поэтому асинхронная обработка связана с исходным запросом. В структурированных логах операций есть `trace_id`, `span_id`, `correlation_id`, имя сервиса и экземпляра.

1. Откройте Grafana → Explore → Loki.
2. Найдите запрос, например: `{app="airline_tracker"} | json | correlation_id="my-request-1"`.
3. Раскройте строку и нажмите `Open trace` у поля `TraceID`.
4. В Tempo откройте операции API, publisher и notifier; из span можно перейти к связанным логам.

Трассу можно найти непосредственно в Explore → Tempo по Trace ID или запросом TraceQL `{ resource.service.name = "publisher" }`. Для уведомлений, которые ещё ожидают времени отправки, SMTP span появится после выполнения задания.

Мониторинг использует ограниченную очередь экспорта. Недоступность Alloy или Tempo не останавливает бизнес-операции; при заполнении очереди часть телеметрии может быть потеряна.

## Экземпляры и панели

В основном Compose запускаются три publisher (`api-publisher`, `api-publisher-2`, `api-publisher-3`) и три notifier (`notifier`, `notifier-2`, `notifier-3`). Для каждого задан свой `INSTANCE_ID`. Notifier используют одну Kafka consumer group: partitions распределяются между ними. Publisher и отправители писем распределяют задания через PostgreSQL. Эти экземпляры подключаются к primary для изменения состояния очередей; читающие реплики БД не используются для получения заданий.

Панель **Airline Tracker — Workers** показывает доступность экземпляров, частоту операций, повторы и ошибки, время выполнения, размер и возраст очередей, а также логи. Список экземпляров можно ограничить фильтром `Worker instance`.

Размер очереди общий для нескольких workers, поэтому панель берёт максимум показаний, а не складывает их. Метрика `up` показывает доступность HTTP endpoint метрик; успешную обработку заданий нужно оценивать по счётчикам операций и состоянию очереди.

## Настройки

| Настройка | Значение по умолчанию | Назначение |
| --- | --- | --- |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://alloy:4318` в основном Compose | Базовый OTLP/HTTP endpoint; exporter добавляет `/v1/traces` |
| `OTEL_TRACES_SAMPLER_ARG` | `1` | Доля новых корневых трасс от 0 до 1; дочерние учитывают решение родителя |
| `INSTANCE_ID` | Имя сервиса Compose | Отличает рабочие экземпляры в логах и трассировках |
| `API_REQUEST_TIMEOUT` | `10s` | Бюджет обработки HTTP-запроса; значение должно быть положительной длительностью |

Без `OTEL_EXPORTER_OTLP_ENDPOINT` при отдельном запуске Go-процесса контекст создаётся и передаётся, но spans не экспортируются. В локальном стенде для проверки всей цепочки sampling должен быть `1`.

API ограничивает чтение заголовков пятью секундами, чтение запроса — десятью, idle-соединение — минутой. Общий бюджет обработки отменяет request context и возвращает HTTP 503 при истечении; прерванная операция записи могла уже зафиксироваться в БД, поэтому повтор требует учёта идемпотентности. Nginx использует connect timeout 3s и send/read timeout 15s для API, автоматические повторы upstream отключены. При увеличении `API_REQUEST_TIMEOUT` нужно согласованно изменить ограничения Nginx.

PostgreSQL при подключении получает `statement_timeout=10000`, `lock_timeout=3000`, `idle_in_transaction_session_timeout=10000` (миллисекунды); эти значения можно переопределить в `POSTGRES_URI`. Подключение ограничено пятью секундами, начальная проверка — общим бюджетом 15s и тремя попытками. Redis учитывает context, имеет connect/read/write/pool timeout 3s и не более двух внутренних повторов.

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

Tempo проверяется запуском в изолированном проекте и проверкой `/ready` внутри сети Docker. Конфигурация находится в `backend/monitoring/tempo.yml`; datasource и переходы между логами и traces задаются в `backend/monitoring/grafana/provisioning/datasources/datasources.yml`.

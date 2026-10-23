# PostgreSQL Primary–Replica: две синхронные логические реплики

## Топология

| Узел / сервис | БД | Права |
|---|---|---|
| `postgres`, `api`, `publisher`, `notifier` | Primary | `app_user`, чтение/запись |
| `postgres-replica-1`, `api-read-1`, `api-mirror` | Replica 1 | `app_reader`, чтение |
| `postgres-replica-2`, `api-read-2`, `api-mirror-read-1`, `api-mirror-read-2` | Replica 2 | `app_reader`, чтение |

На репликах `app_reader` имеет `default_transaction_read_only=on`. Администратор `postgres` и внутренний apply worker могут изменять данные -- без этого логическая репликация не работает. Автоматического failover и повышения реплик до primary нет.

## Синхронность и задержка

На primary заданы:

```conf
wal_level = logical
synchronous_commit = remote_apply
synchronous_standby_names = 'FIRST 2 (airline_replica_1, airline_replica_2)'
```

Подписки подключаются с соответствующими `application_name`, именами слотов и `synchronous_commit='on'`. Primary подтверждает COMMIT после применения на обеих репликах.

## Файлы и окружение

- `backend/db/replication/primary.conf`, `replica.conf`, `pg_hba.conf` -- серверные настройки и парольная аутентификация
- `primary-entrypoint.sh` позволяет initdb выполнить CREATE DATABASE и начальные роли с локальным подтверждением, пока подписчиков ещё нет
- У API-мигратора и notifier-мигратора в Compose задано `PGOPTIONS=-c synchronous_commit=local`
- `replication.env` содержит административное подключение и пароль отдельного `replication_user`

## Порядок инициализации

1. Primary и оба подписчика инициализируют собственные тома и роли. Healthcheck проверяет TCP-сервер
2. API-миграции и затем notifier-миграции подготавливают схему primary
3. `replication-setup` создаёт ограниченную роль репликации и публикацию обычных таблиц `api_schema`, `notifier_schema`, `replication_meta` (Goose исключён из публикации)
4. На пустых подписчиках импортируется schema-only dump: таблицы, функции, ограничения и права
5. Создаются две подписки с начальным `copy_data=true`. Скрипт ждёт, пока все опубликованные таблицы имеют `srsubstate='r'`, и пока обе разные реплики имеют `state='streaming'`, `sync_state='sync'`.
6. После завершения начального копирования и подключения обеих синхронных реплик API, publisher, notifier и затем nginx допускаются к запуску. Диагностическая запись, проверка запрета записи и вывод задержек выполняются отдельно через `make backend-replication-check`.

## Обновление существующих томов

При добавлении полей outbox и уведомлений `replication-setup` применяет schema-only части соответствующих миграций к уже существующим подписчикам перед обновлением публикации. Новые поля и таблица inbox должны существовать на каждой реплике: логическая репликация не передаёт DDL. Для новых томов по-прежнему используется полный schema-only dump primary.

Перенос старых уведомлений на отдельные задания для получателей выполняется отдельной миграцией данных только на primary. Её изменения доставляются через подписки; повторять backfill вручную на репликах нельзя.

Обновление выполняется с остановленными прикладными процессами: сначала остановите API, publisher и notifier (включая их дополнительные экземпляры), затем выполните `make backend-release`. Compose запускает миграции и `replication-setup` до новых приложений. Существующие тома сохраняются; `make down` для такого обновления не подходит, поскольку удаляет тома.

## Проверка

`make backend-replication-check`:
- проверяет синхронные настройки и подключение обеих реплик
- обновляет служебную строку `replication_meta.probe` и после успешного COMMIT проверяет её значение на подписчиках
- подключается как `app_reader` к каждой реплике и проверяет read-only режим
- проверяет отказ `UPDATE` в обычной транзакции и после `BEGIN READ WRITE`
- использует ROLLBACK, чтобы ошибочная выдача прав не оставила изменение при тесте отказа
- выводит `pg_stat_subscription` и `pg_stat_replication`, включая `LSN`, `write_lag`, `flush_lag`, `replay_lag` и расстояние replay в байтах

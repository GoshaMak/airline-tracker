# Тестовое окружение

Стенд содержит отдельные PostgreSQL 16, Redis 7, Kafka, Mailpit, API, publisher и notifier.

## Управление окружением

```sh
make test-env-up
make test-env-check
make test-env-reset
make test-env-down
```

По умолчанию используется `RUN_ID=local`.

`stand.sh` -- набор shell-функций, подключаемый через `.` из корня репозитория.

`environment_up` собирает приложение, поднимает зависимости, при первом запуске применяет миграции и фикстуры, сохраняет снимок БД, создаёт Kafka-топики и запускает приложение. Повторный вызов сохраняет текущие данные. Для их сброса используется `environment_reset`.

## Интеграционные тесты

```sh
make test-env-up RUN_ID=developer-a
make test-integration RUN_ID=developer-a
```

Тесты выделены тегом `integration`, поэтому обычный `go test ./...` их не запускает.

Результаты Allure сохраняются в `tests/environment/results/<RUN_ID>/allure-results/`, лог восстановления -- в `cleanup.log`.

Генерация отчёта:

```sh
allure generate \
  tests/environment/results/developer-a/allure-results \
  --clean -o tests/environment/results/developer-a/allure-report
```

## E2E-тест

```sh
make test-e2e RUN_ID=developer-a
```

Результаты находятся в `results/<RUN_ID>/allure-results/`. Для продолжения работы с приложением выполнить `make test-env-up` с тем же RUN_ID.

## Восстановление исходного состояния

`environment_restore` удаляет оставшиеся контейнеры тестов и останавливает API, publisher и notifier. Утилита `drain` из `cmd/kafka-drain/main.go` ожидает выхода notifier из consumer group, читает оставшиеся сообщения обоих топиков, фиксирует offsets группы `notifier-service` и проверяет `lag=0`. Число прочитанных сообщений и offsets сохраняются в `kafka-drain.jsonl`.

Затем функция восстанавливает снимок PostgreSQL, сравнивает отпечатки всех таблиц схем API и notifier и очищает Redis и Mailpit.

`environment_reset` вызывает общее восстановление и пересоздаёт Kafka с пустыми топиками для следующего прогона. Финальная очистка CI использует то же восстановление, затем удаляет стенд.

После сброса приложение остаётся остановленным. Для продолжения вызвать `environment_up`.

## Удаление окружения

`environment_down` удаляет контейнеры, сеть и тома выбранного стенда.

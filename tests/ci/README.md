# CI в GitHub Actions

## Отчётность

В артефакте `test-results-<run_id>-<attempt>` находятся:

- `allure-report/index.html` -- готовый самостоятельный HTML-отчёт
- `allure-results/` -- исходные результаты тестов
- `commit.txt`, `tools.txt`, `*.status` -- commit, версии инструментов и статусы этапов
- логи этапов, `services.log`, `coverage.out`, `coverage.html`
- `final-cleanup.log`, `kafka-drain.jsonl` и отпечатки БД -- результаты восстановления

Чтобы посмотреть отчёт, скачайте артефакт на странице запуска GitHub Actions,
распакуйте архив и откройте `allure-report/index.html` в браузере.

CI сначала формирует обычный отчёт для сохранения истории запусков, затем заменяет
его однофайловым. История остаётся в `tests/ci/history` и сохраняется в кэше Actions.
Отчёт формируется и после падения тестов, незапущенные этапы отмечаются как `skipped`.

## Локальный запуск

```sh
make test-ci RUN_ID=developer-a-1
```

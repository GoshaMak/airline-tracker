# CI в GitHub Actions

## Отчётность

В артефакте `test-results-<run_id>-<attempt>` находятся:

- `allure-report/`, `allure-results/` -- HTML-отчёт и результаты
- `commit.txt`, `tools.txt`, `*.status` -- commit, версии инструментов и статусы этапов
- логи этапов, `services.log`, `coverage.out`, `coverage.html`
- `final-cleanup.log`, `kafka-drain.jsonl` и отпечатки БД -- результаты восстановления

Открыть отчёт:

```bash
allure open <путь>/allure-report
```

## Локальный запуск

```sh
make test-ci RUN_ID=developer-a-1
```

import fcntl
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import time
import uuid


def prepare_statuses(results):
    statuses = {}
    for stage in ("checkout", "prepare", "unit", "integration", "e2e", "cleanup"):
        path = results / f"{stage}.status"
        state = path.read_text().strip() if path.exists() else "pending"
        statuses[stage] = {"pending": "skipped", "running": "failed"}.get(state, state)
        path.write_text(statuses[stage] + "\n")
    return statuses


def add_missing_results(raw, statuses):
    raw.mkdir(parents=True, exist_ok=True)
    existing = [json.loads(path.read_text()) for path in raw.glob("*-result.json")]
    for stage, state in statuses.items():
        failed_test = any(
            test.get("status") in ("failed", "broken")
            and {"name": "epic", "value": stage} in test.get("labels", [])
            for test in existing
        )
        if state == "passed" or (state == "failed" and failed_test):
            continue
        now = int(time.time() * 1000)
        name = f"CI stage: {stage}"
        result = {
            "uuid": str(uuid.uuid4()),
            "name": name,
            "fullName": name,
            "historyId": hashlib.md5(name.encode()).hexdigest(),
            "status": state,
            "stage": "finished",
            "start": now,
            "stop": now,
            "statusDetails": {
                "message": (
                    "Stage was not executed"
                    if state == "skipped"
                    else "Stage failed or was interrupted; see its log"
                ),
            },
            "labels": [
                {"name": "epic", "value": stage},
                {"name": "suite", "value": "CI stages"},
            ],
        }
        (raw / f"ci-stage-{stage}-result.json").write_text(json.dumps(result))


def generate_report(results, history):
    raw = results / "allure-results"
    history.mkdir(parents=True, exist_ok=True)
    with (history / ".lock").open("w") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        trend = history / "history-trend.json"
        previous = json.loads(trend.read_text()) if trend.exists() else []
        order = max((run.get("buildOrder") or 0 for run in previous), default=0) + 1
        (raw / "history").mkdir(exist_ok=True)
        for path in history.glob("*.json"):
            shutil.copy2(path, raw / "history" / path.name)

        run_url = ""
        if os.environ.get("GITHUB_RUN_ID"):
            run_url = f"{os.environ.get('GITHUB_SERVER_URL', 'https://github.com')}/{os.environ['GITHUB_REPOSITORY']}/actions/runs/{os.environ['GITHUB_RUN_ID']}"
        executor = {
            "name": "GitHub Actions" if run_url else "Local CI",
            "type": "github",
            "buildOrder": order,
            "buildName": os.environ["CI_RUN_ID"],
            "buildUrl": run_url,
            "reportUrl": run_url,
            "reportName": "Airline Tracker tests",
        }
        (raw / "executor.json").write_text(json.dumps(executor))
        command = [
            "allure",
            "generate",
            str(raw),
            "--clean",
            "-o",
            str(results / "allure-report"),
        ]
        subprocess.run(command, check=True)
        shutil.copytree(results / "allure-report/history", history, dirs_exist_ok=True)
        subprocess.run(command + ["--single-file"], check=True)


if __name__ == "__main__":
    results = Path(os.environ["CI_RESULTS"])
    results.mkdir(parents=True, exist_ok=True)
    try:
        statuses = prepare_statuses(results)
        add_missing_results(results / "allure-results", statuses)
        generate_report(results, Path(os.environ["CI_HISTORY"]))
    finally:
        subprocess.run(["chmod", "-R", "a+rX", str(results)], check=True)

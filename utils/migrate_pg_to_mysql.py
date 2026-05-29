import json
import os
from contextlib import closing
from decimal import Decimal

import psycopg2
import pymysql


TABLES = [
    ("users", ["id", "email", "password_hash", "role"]),
    ("airports", ["id", "iata_code", "title", "city", "country"]),
    ("gates", ["id", "airport_id", "number"]),
    (
        "aircraft_models",
        ["id", "manufacturer", "model", "mass", "max_altitude", "max_speed"],
    ),
    (
        "aircraft",
        ["id", "aircraft_model_id", "registration_number", "serial_number", "mileage"],
    ),
    (
        "flights",
        [
            "id",
            "aircraft_id",
            "scheduled_departure",
            "scheduled_arrival",
            "actual_departure",
            "actual_arrival",
            "status",
            "plan",
        ],
    ),
    ("flight_routes", ["id", "flight_id", "departure_gate_id", "arrival_gate_id"]),
    ("subscriptions", ["id", "user_id", "flight_id"]),
    ("outbox", ["id", "topic", "payload", "created_at", "sent_at"]),
    (
        "notifications",
        ["id", "payload", "created_at", "send_at", "status", "type"],
    ),
]

BATCH_SIZE = int(os.getenv("MIGRATION_BATCH_SIZE", "1000"))


def env(*names: str, default: str | None = None) -> str | None:
    for name in names:
        value = os.getenv(name)
        if value:
            return value
    return default


def connect_postgres():
    return psycopg2.connect(
        dbname=env("POSTGRES_DB", "DB_NAME", default="airline_tracker"),
        user=env("POSTGRES_USER", "DB_USER", default="postgres"),
        password=env("POSTGRES_PASSWORD", "DB_PASSWORD", default="postgres"),
        host=env("POSTGRES_HOST", "DB_HOST", default="localhost"),
        port=env("POSTGRES_PORT", "DB_PORT", default="5432"),
    )


def connect_mysql():
    return pymysql.connect(
        database=env("MYSQL_DATABASE", "MYSQL_NAME", default="airline_tracker"),
        user=env("MYSQL_USER", default="app_user"),
        password=env("MYSQL_PASSWORD", default="app_pswd"),
        host=env("MYSQL_HOST", default="localhost"),
        port=int(env("MYSQL_PORT", default="3306")),
        charset="utf8mb4",
        autocommit=False,
    )


def normalize(value):
    if isinstance(value, dict | list):
        return json.dumps(value, separators=(",", ":"), ensure_ascii=False)
    if isinstance(value, Decimal):
        return int(value) if value == value.to_integral_value() else float(value)
    if isinstance(value, memoryview):
        return bytes(value)
    return value


def quoted_columns(columns: list[str]) -> str:
    return ", ".join(f"`{column}`" for column in columns)


def select_sql(table: str, columns: list[str]) -> str:
    return f"select {', '.join(columns)} from {table}"


def upsert_sql(table: str, columns: list[str]) -> str:
    placeholders = ", ".join(["%s"] * len(columns))
    updates = ", ".join(
        f"`{column}` = values(`{column}`)" for column in columns if column != "id"
    )
    if not updates:
        updates = "`id` = values(`id`)"
    return (
        f"insert into `{table}` ({quoted_columns(columns)}) "
        f"values ({placeholders}) "
        f"on duplicate key update {updates}"
    )


def copy_table(pg_cursor, mysql_cursor, table: str, columns: list[str]) -> int:
    pg_cursor.execute(select_sql(table, columns))
    insert = upsert_sql(table, columns)
    copied = 0

    while True:
        rows = pg_cursor.fetchmany(BATCH_SIZE)
        if not rows:
            break

        normalized_rows = [tuple(normalize(value) for value in row) for row in rows]
        mysql_cursor.executemany(insert, normalized_rows)
        copied += len(normalized_rows)

    return copied


def main():
    with closing(connect_postgres()) as pg_conn, closing(connect_mysql()) as mysql_conn:
        with closing(pg_conn.cursor()) as pg_cursor, closing(
            mysql_conn.cursor()
        ) as mysql_cursor:
            mysql_cursor.execute("set foreign_key_checks = 0")
            try:
                for table, columns in TABLES:
                    copied = copy_table(pg_cursor, mysql_cursor, table, columns)
                    mysql_conn.commit()
                    print(f"{table}: copied {copied} rows")
            except Exception:
                mysql_conn.rollback()
                raise
            finally:
                mysql_cursor.execute("set foreign_key_checks = 1")
                mysql_conn.commit()

    print("migration finished successfully")


if __name__ == "__main__":
    main()

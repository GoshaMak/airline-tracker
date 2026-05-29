import datetime as dt
import decimal
import json
import os
import uuid

import psycopg2
from psycopg2.extras import RealDictCursor
from pymongo import ASCENDING, MongoClient, ReplaceOne
from pymongo.errors import BulkWriteError, OperationFailure


PG_CONFIG = {
    "dbname": os.getenv("POSTGRES_NAME")
    or os.getenv("POSTGRES_DB")
    or os.getenv("DB_NAME")
    or "airline_tracker",
    "user": os.getenv("POSTGRES_USER", "postgres"),
    "password": os.getenv("POSTGRES_PASSWORD", "postgres"),
    "host": os.getenv("POSTGRES_HOST", "postgres"),
    "port": os.getenv("POSTGRES_PORT", "5432"),
}

MONGO_HOST = os.getenv("MONGO_HOST", "localhost")
MONGO_PORT = os.getenv("MONGO_PORT", "27017")
MONGO_USER = os.getenv("MONGO_USER", "mongo")
MONGO_PASSWORD = os.getenv("MONGO_PASSWORD", "mongo")
MONGO_DB_NAME = (
    os.getenv("MONGO_NAME")
    or os.getenv("MONGO_DB")
    or os.getenv("DB_NAME")
    or "airline_tracker"
)

if MONGO_USER and MONGO_PASSWORD:
    MONGO_URI = (
        f"mongodb://{MONGO_USER}:{MONGO_PASSWORD}@{MONGO_HOST}:{MONGO_PORT}/"
        f"{MONGO_DB_NAME}?authSource=admin"
    )
else:
    MONGO_URI = f"mongodb://{MONGO_HOST}:{MONGO_PORT}"


TABLE_QUERIES = {
    "users": "select * from users",
    "airports": "select * from airports",
    "gates": "select * from gates",
    "aircraft_models": "select * from aircraft_models",
    "aircraft": "select * from aircraft",
    "flights": """
        select
            f.id,
            f.aircraft_id,
            f.scheduled_departure,
            f.scheduled_arrival,
            f.actual_departure,
            f.actual_arrival,
            f.status,
            f.plan,
            fr.departure_gate_id,
            fr.arrival_gate_id,
            gd.airport_id as departure_airport_id,
            ga.airport_id as arrival_airport_id
        from flights f
            join flight_routes fr on fr.flight_id = f.id
            join gates gd on gd.id = fr.departure_gate_id
            join gates ga on ga.id = fr.arrival_gate_id
    """,
    "flight_routes": "select * from flight_routes",
    "subscriptions": "select * from subscriptions",
    "outbox": "select * from outbox",
    "notifications": "select * from notifications",
}


def convert_value(value):
    if isinstance(value, uuid.UUID):
        return str(value)
    if isinstance(value, decimal.Decimal):
        return int(value) if value == value.to_integral_value() else float(value)
    if isinstance(value, memoryview):
        return bytes(value)
    if isinstance(value, list):
        return [convert_value(item) for item in value]
    if isinstance(value, dict):
        return {key: convert_value(item) for key, item in value.items()}
    if isinstance(value, dt.datetime):
        return value
    return value


def convert_row(table, row):
    doc = {key: convert_value(value) for key, value in row.items()}

    if table == "notifications" and isinstance(doc.get("payload"), (dict, list)):
        doc["payload"] = json.dumps(doc["payload"], separators=(",", ":")).encode()

    if "id" in doc:
        doc["_id"] = doc.pop("id")

    return doc


def ensure_indexes(db):
    indexes = [
        ("users", [("email", ASCENDING)], True),
        ("airports", [("iata_code", ASCENDING)], True),
        ("airports", [("title", ASCENDING)], True),
        ("gates", [("airport_id", ASCENDING), ("number", ASCENDING)], True),
        (
            "aircraft_models",
            [("manufacturer", ASCENDING), ("model", ASCENDING)],
            True,
        ),
        ("aircraft", [("registration_number", ASCENDING)], True),
        ("aircraft", [("serial_number", ASCENDING)], True),
        ("flight_routes", [("flight_id", ASCENDING)], True),
        ("subscriptions", [("user_id", ASCENDING), ("flight_id", ASCENDING)], True),
        ("outbox", [("sent_at", ASCENDING)], False),
        ("notifications", [("status", ASCENDING)], False),
    ]

    for collection_name, keys, unique in indexes:
        try:
            db[collection_name].create_index(keys, unique=unique)
        except OperationFailure as exc:
            print(f"  Индекс {collection_name} {keys} не создан: {exc}")


def migrate_collection(pg_cur, db, table, query):
    print(f"\nМиграция таблицы: {table}")
    pg_cur.execute(query)
    rows = pg_cur.fetchall()
    db[table].delete_many({})

    if not rows:
        print(f"  Таблица {table} пуста. Коллекция очищена.")
        return

    docs = [convert_row(table, dict(row)) for row in rows]
    operations = [
        ReplaceOne({"_id": doc["_id"]}, doc, upsert=True)
        for doc in docs
        if "_id" in doc
    ]

    if not operations:
        print(f"  Нет документов с id для миграции.")
        return

    try:
        result = db[table].bulk_write(operations, ordered=False)
        print(
            "  Обработано: "
            f"{len(operations)}, вставлено: {result.upserted_count}, "
            f"обновлено: {result.modified_count}."
        )
    except BulkWriteError as exc:
        details = exc.details
        print(
            "  Ошибка bulk write: "
            f"inserted={details.get('nInserted', 0)}, "
            f"upserted={details.get('nUpserted', 0)}, "
            f"modified={details.get('nModified', 0)}, "
            f"errors={len(details.get('writeErrors', []))}"
        )
        raise


def migrate():
    print(f"Подключение к PostgreSQL ({PG_CONFIG['host']}:{PG_CONFIG['port']})...")
    pg_conn = psycopg2.connect(**PG_CONFIG)
    pg_cur = pg_conn.cursor(cursor_factory=RealDictCursor)

    print(f"Подключение к MongoDB ({MONGO_HOST}:{MONGO_PORT}/{MONGO_DB_NAME})...")
    mongo_client = MongoClient(MONGO_URI)
    mongo_client.admin.command("ping")
    db = mongo_client[MONGO_DB_NAME]

    try:
        ensure_indexes(db)
        for table, query in TABLE_QUERIES.items():
            migrate_collection(pg_cur, db, table, query)
    finally:
        pg_cur.close()
        pg_conn.close()
        mongo_client.close()
        print("\nМиграция завершена.")


if __name__ == "__main__":
    migrate()

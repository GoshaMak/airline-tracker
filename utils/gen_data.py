import os
import random
import uuid
from datetime import timedelta, timezone
from datetime import datetime

import psycopg2
from faker import Faker

USERS_AMT = 100
USER_ROLES = ["admin", "user"]
AIRPORTS_AMT = 100
UPPER_CASE_LETTERS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
AIRCRAFT_MODELS = [
    "Airbus",
    "Boeing",
    "Embraer",
    "Antonov",
    "ATR",
    "Bombardier",
    "Cessna",
    "COMAC",
    "Convair",
    "Dassault",
    "Fokker",
    "Ilyushin",
    "Lockheed",
    "Mitsubishi",
    "Saab",
    "Sukhoi",
    "Tupo",
]
MASS_MIN = 1
MASS_MAX = 1_000_000
ALTIDUDE_MIN = 1
ALTIDUDE_MAX = 20_000
SPEED_MIN = 1
SPEED_MAX = 10_000
AIRCRAFTS_AMT = 10_000
MILEAGE_MIN = 0
MILEAGE_MAX = 50_000
FLIGHT_STATUSES = [
    "scheduled",
    "boarding",
    "departed",
    "landed",
    "arrived",
    "delayed",
    "cancelled",
    "rescheduled",
]

users = []
airports = []
gates = []
gateToAirport = {}
aircraftModels = []
aircrafts = []
flights = []
flightRoutes = []
subscriptions = []

fake = Faker()
conn = psycopg2.connect(
    dbname=os.getenv("POSTGRES_DB"),
    user=os.getenv("POSTGRES_USER"),
    password=os.getenv("POSTGRES_PASSWORD"),
    host=os.getenv("POSTGRES_HOST"),
    port=os.getenv("POSTGRES_PORT"),
)
cur = conn.cursor()


def ResetDatabase():
    cur.execute(
        """
        truncate table
            notifications,
            outbox,
            subscriptions,
            flight_routes,
            flights,
            aircraft,
            gates,
            aircraft_models,
            airports,
            users
        restart identity cascade
        """
    )


def genRandomInt(min: int, max: int) -> int:
    return random.randint(min, max)


def genRandomRepresentation() -> str:
    reprs = ["alpha-2", "alpha-3"]
    return random.choice(reprs)


def genGateNumber() -> str:
    return fake.bothify(text="?%#!", letters=UPPER_CASE_LETTERS)


def GenUsers():
    for _ in range(USERS_AMT):
        user = (
            str(uuid.uuid4()),
            fake.unique.email(),
            # fake.password(
            #     length=genRandomInt(8, 100),
            #     special_chars=True,
            #     digits=True,
            #     upper_case=True,
            #     lower_case=True,
            # ),
            fake.sha256()[:60],
            "user",
        )
        users.append(user)

    cur.executemany(
        """insert into users 
           (id, email, password_hash, role) 
           values (%s, %s, %s, %s)""",
        users,
    )


def GenAirports():
    for _ in range(AIRPORTS_AMT):
        fake.lexify()
        airport = (
            str(uuid.uuid4()),
            fake.unique.lexify(text="???", letters=UPPER_CASE_LETTERS).upper(),  # IATA
            fake.unique.company() + " Airport",
            fake.city(),
            fake.country_code(representation=genRandomRepresentation()),
        )
        airports.append(airport)

    cur.executemany(
        "insert into airports (id, iata_code, title, city, country) values (%s, %s, %s, %s, %s)",
        airports,
    )


def GenGates():
    for airport in airports:
        airportId = airport[0]
        gateNumbers = []
        amt = genRandomInt(1, 10)  # gates amt per airport
        for _ in range(amt):
            gateId = str(uuid.uuid4())
            gateNumber = genGateNumber()
            while gateNumber in gateNumbers:
                gateNumber = genGateNumber()
            gateNumbers.append(gateNumber)
            gates.append((gateId, airportId, gateNumber))
            gateToAirport[gateId] = airportId

    cur.executemany(
        "insert into gates (id, airport_id, number) values (%s, %s, %s)",
        gates,
    )


def GenAircraftModels():
    for i in range(len(AIRCRAFT_MODELS)):
        am = AIRCRAFT_MODELS[i]
        ams = []
        for _ in range(random.randint(1, 4)):
            tmpM = fake.bothify(text="?%!!!!", letters=UPPER_CASE_LETTERS)
            while tmpM in ams:
                tmpM = fake.bothify(text="?%!!!!", letters=UPPER_CASE_LETTERS)
            ams.append(tmpM)
            aircraftModels.append(
                (
                    str(uuid.uuid4()),
                    am,
                    tmpM,
                    random.randint(MASS_MIN, MASS_MAX),
                    random.randint(ALTIDUDE_MIN, ALTIDUDE_MAX),
                    random.randint(SPEED_MIN, SPEED_MAX),
                )
            )

    cur.executemany(
        """insert into aircraft_models 
           (id, manufacturer, model, mass, max_altitude, max_speed) 
           values (%s, %s, %s, %s, %s, %s)""",
        aircraftModels,
    )


def GenAircrafts():
    for i in range(len(aircraftModels)):
        aircraftModel = aircraftModels[i]
        aircraftModelId = aircraftModel[0]
        for _ in range(random.randint(5, 10)):
            aircraftId = str(uuid.uuid4())
            fake.bothify()
            aircrafts.append(
                (
                    aircraftId,
                    aircraftModelId,
                    fake.unique.bothify(text="?!-###!!", letters=UPPER_CASE_LETTERS),
                    fake.unique.bothify(text="??-##!!!!!", letters=UPPER_CASE_LETTERS),
                    random.randint(MILEAGE_MIN, MILEAGE_MAX),
                )
            )

    cur.executemany(
        """insert into aircraft 
           (id, aircraft_model_id, registration_number, serial_number, mileage)
           values (%s, %s, %s, %s, %s)""",
        aircrafts,
    )


def _GenFlightPlan():
    length = random.randint(0, 200)

    base_chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    extra_chars = "- "

    result = []
    for _ in range(length):
        if random.random() < 0.1:
            result.append(random.choice(extra_chars))
        else:
            result.append(random.choice(base_chars))

    return "".join(result)


def GenFlights():
    amt = int(len(aircrafts) * (random.randint(7, 9) / 10.0))
    now = datetime.now(timezone.utc)

    for i in range(amt):
        aircraft = aircrafts[i]
        flightId = str(uuid.uuid4())
        aircraftId = aircraft[0]

        scenario = random.choices(
            population=[
                "future_scheduled",
                "future_boarding",
                "future_delayed",
                "future_cancelled",
                "future_rescheduled",
                "in_air",
                "landed_not_arrived",
                "completed",
                "past_cancelled",
            ],
            weights=[25, 7, 10, 5, 5, 15, 8, 20, 5],
            k=1,
        )[0]

        status = ""
        schDep, schArr = None, None
        actDep, actArr = None, None

        if scenario == "future_scheduled":
            status = "scheduled"
            schDep = now + timedelta(hours=random.randint(2, 72))
            schArr = schDep + timedelta(hours=random.randint(1, 10))

        elif scenario == "future_boarding":
            status = "boarding"
            schDep = now + timedelta(minutes=random.randint(5, 45))
            schArr = schDep + timedelta(hours=random.randint(1, 10))

        elif scenario == "future_delayed":
            status = "delayed"
            schDep = now + timedelta(minutes=random.randint(30, 360))
            schArr = schDep + timedelta(hours=random.randint(1, 10))

        elif scenario == "future_cancelled":
            status = "cancelled"
            schDep = now - timedelta(hours=random.randint(1, 72))
            schArr = schDep + timedelta(hours=random.randint(1, 10))

        elif scenario == "future_rescheduled":
            status = "rescheduled"
            schDep = now + timedelta(hours=random.randint(2, 96))
            schArr = schDep + timedelta(hours=random.randint(1, 10))

        elif scenario == "in_air":
            status = "departed"
            schDep = now - timedelta(minutes=random.randint(20, 240))
            schArr = now + timedelta(minutes=random.randint(20, 360))

            actDep = schDep + timedelta(minutes=random.randint(-10, 60))
            actArr = None

            if actDep > now:
                actDep = now - timedelta(minutes=random.randint(5, 19))

        elif scenario == "landed_not_arrived":
            status = "landed"
            schDep = now - timedelta(hours=random.randint(2, 12))
            schArr = now - timedelta(minutes=random.randint(5, 60))

            actDep = schDep + timedelta(minutes=random.randint(-10, 60))
            actArr = now - timedelta(minutes=random.randint(1, 30))

        elif scenario == "completed":
            status = "arrived"
            schDep = now - timedelta(hours=random.randint(3, 72))
            schArr = schDep + timedelta(hours=random.randint(1, 10))

            actDep = schDep + timedelta(minutes=random.randint(-10, 60))
            actArr = schArr + timedelta(minutes=random.randint(-20, 90))

            if actArr <= actDep:
                actArr = actDep + timedelta(minutes=random.randint(30, 600))

            if actArr > now:
                actArr = now - timedelta(minutes=random.randint(1, 60))

            if actArr <= actDep:
                actDep = actArr - timedelta(minutes=random.randint(30, 600))

        elif scenario == "past_cancelled":
            status = "cancelled"
            schDep = now - timedelta(hours=random.randint(2, 72))
            schArr = schDep + timedelta(hours=random.randint(1, 10))

        plan = _GenFlightPlan()
        plan = plan.strip()
        if len(plan) < 2:
            plan = None

        flights.append(
            (
                flightId,
                aircraftId,
                schDep,
                schArr,
                actDep,
                actArr,
                status,
                plan,
            )
        )

    cur.executemany(
        """insert into flights 
           (id, aircraft_id, scheduled_departure, scheduled_arrival,
            actual_departure, actual_arrival, status, plan)
           values (%s, %s, %s, %s, %s, %s, %s, %s)""",
        flights,
    )


def GenFlightRoutes():
    for flight in flights:
        flightId = flight[0]

        departureGate = random.choice(gates)
        arrivalGate = random.choice(gates)

        while gateToAirport[arrivalGate[0]] == gateToAirport[departureGate[0]]:
            arrivalGate = random.choice(gates)

        flightRoutes.append(
            (
                str(uuid.uuid4()),
                flightId,
                departureGate[0],
                arrivalGate[0],
            )
        )

    cur.executemany(
        """insert into flight_routes
           (id, flight_id, departure_gate_id, arrival_gate_id)
           values (%s, %s, %s, %s)""",
        flightRoutes,
    )


def GenSubscriptions():
    usedPairs = set()

    amt = min(len(users) * 3, len(users) * len(flights))

    for _ in range(amt):
        userId = random.choice(users)[0]
        flightId = random.choice(flights)[0]

        while (userId, flightId) in usedPairs:
            userId = random.choice(users)[0]
            flightId = random.choice(flights)[0]

        usedPairs.add((userId, flightId))

        subscriptions.append(
            (
                str(uuid.uuid4()),
                userId,
                flightId,
            )
        )

    cur.executemany(
        """insert into subscriptions
           (id, user_id, flight_id)
           values (%s, %s, %s)""",
        subscriptions,
    )


if __name__ == "__main__":
    ResetDatabase()
    print("Database reset successfully")
    GenUsers()
    print("Users filled successfully")
    GenAirports()
    print("Airports filled successfully")
    GenGates()
    print("Gates filled successfully")
    GenAircraftModels()
    print("AircraftModels filled successfully")
    GenAircrafts()
    print("Aircrafts filled successfully")
    GenFlights()
    print("Flights filled successfully")
    GenFlightRoutes()
    print("FlightRoutes filled successfully")
    GenSubscriptions()
    print("Subscriptions filled successfully")

    conn.commit()

    cur.close()
    conn.close()

    print("Database filled successfully")

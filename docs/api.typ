#set page(
  paper: "a4",
  margin: (x: 0.5cm, y: 1cm),
)

#set text(
  font: "Noto Sans Adlam",
  size: 14pt,
)

#set table(
  align: (x, y) => if y == 0 {
    { center }
  } else {
    { auto }
  },
)

#show table.cell: it => {
  let (x, y) = (it.x, it.y)
  let alignment = left

  if it.body == [+] or it.body == [-] {
    alignment = center
  }
  alignment += horizon
  align(alignment)[#it]
}

= API endpoints

#table(
  columns: 6,
  table.header([METHOD], [PATH], [PROGRESS], [DESC], [REQ], [RESP]),

  table.cell(
    [GET],
    rowspan: 4,
    fill: blue,
  ),
  [/api/v1/status], [+], [#align(center + horizon)[status]], [-], [str],

  [/api/v1/subscriptions], [+], [user's subscribed flights],
  ["Authorization": "Bearer ..."], [{"flights": [...]}],

  [/api/v1/flights], [+], [get all flights], [-], [{"flights": [...]}],
  [/api/v1/flights/{id}], [+], [get flight by id], [-], [{"flight": Flight}],

  table.cell(
    [POST],
    rowspan: 9,
    fill: green,
  ),
  [/api/v1/users], [+], [register user],
  [{"email": str, "password": str}], [{"msg": str}],

  [/api/v1/auth/tokens], [+], [authenticate user],
  [{"email": str, "password": str}], [{"token": str}],

  [/api/v1/subscriptions], [+], [subscribe to flight's updates],
  [{"token": "Bearer ...", "flight_id": uuid}], [{"msg": str}],

  [/api/v1/uuids], [+], [generate UUID], [-], [uuid],

  [/api/v1/flights], [+], [add flight],
  [{"token": "Bearer ...", "flight": Flight, "aircraft": Aircraft,
    "departure_airport": Airport, "arrival_airport": Airport,
    "departure_gate": Gate, "arrival_gate": Gate}],
  [{"msg": str}],

  [/api/v1/aircraft], [+], [add aircraft],
  [{"token": "Bearer ...", "aircraft": Aircraft}], [{"msg": str}],

  [/api/v1/aircraft-models], [+], [add aircraft model],
  [{"token": "Bearer ...", "aircraft_model": AircraftModel}], [{"msg": str}],

  [/api/v1/airports], [+], [add airport],
  [{"token": "Bearer ...", "airport": Airport}], [{"msg": str}],

  [/api/v1/gates], [+], [add gate],
  [{"token": "Bearer ...", "gate": Gate}], [{"msg": str}],

  table.cell(
    [PATCH],
    rowspan: 1,
    fill: orange,
  ),
  [/api/v1/flights/{id}], [+], [update flight by id],
  [{"token": "Bearer ...", "flight_id": int, "flight": Flight}], [{"msg": str}],
)

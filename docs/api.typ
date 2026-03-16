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
  [/ping], [+], [#align(center + horizon)[ping]], [-], [{"msg": "pong"}],

  [/user/flights], [+], [user's subscribed flights],
  ["Authorization": "Bearer ..."], [{"flights": [...]}],

  [/flights], [+], [get all flights], [-], [{"flights": [...]}],
  [/flight/{id}], [+], [get flight by id], [-], [{"flight": Flight}],

  table.cell(
    [POST],
    rowspan: 9,
    fill: green,
  ),
  [/auth/signup], [+], [-],
  [{"email": str, "phone": str, "password": str, "role": str}], [{"msg": str}],

  [/auth/login], [+], [-],
  [{"email": str, "phone": str, "password": str}], [{"token": str}],

  [/user/subscribe/flight/{id}], [-], [subscribe to flight's updates],
  [{"token": "Bearer ...", "flight_id": int}], [{"msg": str}],

  [/user/unsubscribe/flight/{id}], [-], [unsubscribe from flight's updates],
  [{"token": "Bearer ...", "flight_id": int}], [{"msg": str}],

  [/admin/flight], [+], [add flight],
  [{"token": "Bearer ...", "flight": Flight, "aircraft": Aircraft,
    "departure_airport": Airport, "arrival_airport": Airport,
    "departure_gate": Gate, "arrival_gate": Gate}],
  [{"msg": str}],

  [/admin/aircraft], [+], [add aircraft],
  [{"token": "Bearer ...", "aircraft": Aircraft}], [{"msg": str}],

  [/admin/aircraft_model], [+], [add aircraft model],
  [{"token": "Bearer ...", "aircraft_model": AircraftModel}], [{"msg": str}],

  [/admin/airport], [+], [add airport],
  [{"token": "Bearer ...", "airport": Airport}], [{"msg": str}],

  [/admin/gate], [+], [add gate],
  [{"token": "Bearer ...", "gate": Gate}], [{"msg": str}],

  table.cell(
    [PATCH],
    rowspan: 1,
    fill: orange,
  ),
  [/admin/flight/{id}], [+], [update flight by id],
  [{"token": "Bearer ...", "flight_id": int, "flight": Flight}], [{"msg": str}],

  table.cell(
    [DELETE],
    rowspan: 1,
    fill: red,
  ),
  [/admin/flight/{id}], [+], [delete flight by id],
  [{"token": "Bearer ...", "flight_id": int}], [{"msg": str}],
)

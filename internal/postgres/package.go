package postgres

import "github.com/samber/do/v2"

var Package = do.Package(
	do.Lazy(NewUserRepository),
	do.Lazy(NewFlightRepository),
	do.Lazy(NewAircraftRepository),
	do.Lazy(NewAirportRepository),
	do.Lazy(NewGateRepository),
	do.Lazy(NewAircraftModelRepository),
)

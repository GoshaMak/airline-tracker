package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	USERS_AMT    = 1000
	AIRPORTS_AMT = 10000

	UPPER_CASE_LETTERS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	AIRCRAFTS_AMT = 10000

	MASS_MIN = 1
	MASS_MAX = 1_000_000

	ALTITUDE_MIN = 1
	ALTITUDE_MAX = 20_000

	SPEED_MIN = 1
	SPEED_MAX = 10_000

	MILEAGE_MIN = 0
	MILEAGE_MAX = 50_000
)

var (
	FLIGHTS_AMT int

	AIRCRAFT_MODELS = []string{
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
	}

	users          []User
	countries      []Country
	cities         []City
	airports       []Airport
	gates          []Gate
	gateToAirport  = make(map[string]string)
	aircraftModels []AircraftModel
	aircrafts      []Aircraft
	flights        []Flight
	flightRoutes   []FlightRoute
	subscriptions  []Subscription
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
}

type Country struct {
	ID   string
	Code string
	Name string
}

type City struct {
	ID        string
	Name      string
	CountryID string
}

type Airport struct {
	ID       string
	IATACode string
	Title    string
	CityID   string
}

type Gate struct {
	ID        string
	AirportID string
	Number    string
}

type AircraftModel struct {
	ID           string
	Manufacturer string
	Model        string
	Mass         int
	MaxAltitude  int
	MaxSpeed     int
}

type Aircraft struct {
	ID                 string
	AircraftModelID    string
	RegistrationNumber string
	SerialNumber       string
	Mileage            int
}

type Flight struct {
	ID                 string
	AircraftID         string
	ScheduledDeparture time.Time
	ScheduledArrival   time.Time
	ActualDeparture    *time.Time
	ActualArrival      *time.Time
	Status             string
	Plan               *string
}

type FlightRoute struct {
	ID              string
	FlightID        string
	DepartureGateID string
	ArrivalGateID   string
}

type Subscription struct {
	ID       string
	UserID   string
	FlightID string
}

type CountryFileRow struct {
	Code string
	Name string
}

func Generate(conn *pgxpool.Pool, flights_amt int) {
	ctx := context.Background()

	FLIGHTS_AMT = flights_amt
	resetGeneratedData()

	fmt.Println("Generating database with flights:", FLIGHTS_AMT)

	if err := truncateTables(ctx, conn); err != nil {
		log.Fatalf("truncate tables error: %v", err)
	}

	if err := GenUsers(ctx, conn); err != nil {
		log.Fatalf("GenUsers error: %v", err)
	}
	fmt.Println("Users filled successfully:", len(users))

	if err := GenCountries(ctx, conn); err != nil {
		log.Fatalf("GenCountries error: %v", err)
	}
	fmt.Println("Countries filled successfully:", len(countries))

	if err := GenAirports(ctx, conn); err != nil {
		log.Fatalf("GenAirports error: %v", err)
	}
	fmt.Println("Airports filled successfully:", len(airports))

	if err := GenGates(ctx, conn); err != nil {
		log.Fatalf("GenGates error: %v", err)
	}
	fmt.Println("Gates filled successfully:", len(gates))

	if err := GenAircraftModels(ctx, conn); err != nil {
		log.Fatalf("GenAircraftModels error: %v", err)
	}
	fmt.Println("AircraftModels filled successfully:", len(aircraftModels))

	if err := GenAircrafts(ctx, conn); err != nil {
		log.Fatalf("GenAircrafts error: %v", err)
	}
	fmt.Println("Aircrafts filled successfully:", len(aircrafts))

	if err := GenFlights(ctx, conn); err != nil {
		log.Fatalf("GenFlights error: %v", err)
	}
	fmt.Println("Flights filled successfully:", len(flights))

	if err := GenFlightRoutes(ctx, conn); err != nil {
		log.Fatalf("GenFlightRoutes error: %v", err)
	}
	fmt.Println("FlightRoutes filled successfully:", len(flightRoutes))

	if err := GenSubscriptions(ctx, conn); err != nil {
		log.Fatalf("GenSubscriptions error: %v", err)
	}
	fmt.Println("Subscriptions filled successfully:", len(subscriptions))

	fmt.Println("Database filled successfully")
}

func resetGeneratedData() {
	users = nil
	countries = nil
	cities = nil
	airports = nil
	gates = nil
	gateToAirport = make(map[string]string)
	aircraftModels = nil
	aircrafts = nil
	flights = nil
	flightRoutes = nil
	subscriptions = nil
}

func truncateTables(ctx context.Context, conn *pgxpool.Pool) error {
	query := `
		truncate table
			users,
			countries,
			cities,
			airports,
			gates,
			aircraft_models,
			aircraft,
			flights,
			flight_routes,
			subscriptions,
			outbox,
			notifications
		restart identity cascade;
	`

	_, err := conn.Exec(ctx, query)
	return err
}

func genRandomInt(minVal, maxVal int) int {
	return rand.Intn(maxVal-minVal+1) + minVal
}

func randomChoice[T any](items []T) T {
	return items[rand.Intn(len(items))]
}

func genUUID() string {
	return uuid.NewString()
}

func alphaCode(value, width int) string {
	code := make([]byte, width)
	for i := width - 1; i >= 0; i-- {
		code[i] = UPPER_CASE_LETTERS[value%len(UPPER_CASE_LETTERS)]
		value /= len(UPPER_CASE_LETTERS)
	}
	return string(code)
}

func genPasswordHash() string {
	sum := sha256.Sum256([]byte(gofakeit.Password(true, true, true, true, false, 32)))
	return hex.EncodeToString(sum[:])[:60]
}

func readCountriesFromFile(path string) ([]CountryFileRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []CountryFileRow

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)

		code := strings.ToUpper(strings.TrimSpace(parts[0]))
		name := code

		if len(parts) > 1 {
			name = strings.Join(parts[1:], " ")
		}

		if len(code) != 2 {
			return nil, fmt.Errorf("invalid country code in countries.txt: %s", code)
		}

		result = append(result, CountryFileRow{
			Code: code,
			Name: name,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("countries.txt is empty")
	}

	return result, nil
}

func GenUsers(ctx context.Context, conn *pgxpool.Pool) error {
	for i := 0; i < USERS_AMT; i++ {
		users = append(users, User{
			ID:           genUUID(),
			Email:        fmt.Sprintf("user-%d@example.com", i+1),
			PasswordHash: genPasswordHash(),
			Role:         "user",
		})
	}

	rows := make([][]any, 0, len(users))
	for _, u := range users {
		rows = append(rows, []any{
			u.ID,
			u.Email,
			u.PasswordHash,
			u.Role,
		})
	}

	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"users"},
		[]string{"id", "email", "password_hash", "role"},
		pgx.CopyFromRows(rows),
	)

	return err
}

func GenCountries(ctx context.Context, conn *pgxpool.Pool) error {
	countryRows, err := readCountriesFromFile("./countries.txt")
	if err != nil {
		return err
	}

	usedCodes := make(map[string]struct{})

	for _, row := range countryRows {
		if _, ok := usedCodes[row.Code]; ok {
			continue
		}

		usedCodes[row.Code] = struct{}{}

		countries = append(countries, Country{
			ID:   genUUID(),
			Code: row.Code,
			Name: row.Name,
		})
	}

	rows := make([][]any, 0, len(countries))
	for _, c := range countries {
		rows = append(rows, []any{
			c.ID,
			c.Code,
			c.Name,
		})
	}

	_, err = conn.CopyFrom(
		ctx,
		pgx.Identifier{"countries"},
		[]string{"id", "code", "name"},
		pgx.CopyFromRows(rows),
	)

	return err
}

func GenAirports(ctx context.Context, conn *pgxpool.Pool) error {
	const iataCodeCapacity = 26 * 26 * 26
	if AIRPORTS_AMT > iataCodeCapacity {
		return fmt.Errorf("cannot generate %d unique three-letter IATA codes", AIRPORTS_AMT)
	}

	for i := 0; i < AIRPORTS_AMT; i++ {
		country := randomChoice(countries)
		cityID := genUUID()

		cities = append(cities, City{
			ID:        cityID,
			Name:      fmt.Sprintf("City %d", i+1),
			CountryID: country.ID,
		})

		title := fmt.Sprintf("%s Airport %d", gofakeit.Company(), len(airports)+1)

		airports = append(airports, Airport{
			ID:       genUUID(),
			IATACode: alphaCode(i, 3),
			Title:    title,
			CityID:   cityID,
		})
	}

	cityRows := make([][]any, 0, len(cities))
	for _, c := range cities {
		cityRows = append(cityRows, []any{
			c.ID,
			c.Name,
			c.CountryID,
		})
	}

	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"cities"},
		[]string{"id", "name", "country_id"},
		pgx.CopyFromRows(cityRows),
	)
	if err != nil {
		return err
	}

	airportRows := make([][]any, 0, len(airports))
	for _, a := range airports {
		airportRows = append(airportRows, []any{
			a.ID,
			a.IATACode,
			a.Title,
			a.CityID,
		})
	}

	_, err = conn.CopyFrom(
		ctx,
		pgx.Identifier{"airports"},
		[]string{"id", "iata_code", "title", "city_id"},
		pgx.CopyFromRows(airportRows),
	)

	return err
}

func GenGates(ctx context.Context, conn *pgxpool.Pool) error {
	for _, airport := range airports {
		amt := genRandomInt(1, 10)

		for i := 0; i < amt; i++ {
			gateID := genUUID()

			gates = append(gates, Gate{
				ID:        gateID,
				AirportID: airport.ID,
				Number:    fmt.Sprintf("G%03d", i+1),
			})

			gateToAirport[gateID] = airport.ID
		}
	}

	rows := make([][]any, 0, len(gates))
	for _, g := range gates {
		rows = append(rows, []any{
			g.ID,
			g.AirportID,
			g.Number,
		})
	}

	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"gates"},
		[]string{"id", "airport_id", "number"},
		pgx.CopyFromRows(rows),
	)

	return err
}

func GenAircraftModels(ctx context.Context, conn *pgxpool.Pool) error {
	for _, manufacturer := range AIRCRAFT_MODELS {
		amt := genRandomInt(1, 4)

		for i := 0; i < amt; i++ {
			aircraftModels = append(aircraftModels, AircraftModel{
				ID:           genUUID(),
				Manufacturer: manufacturer,
				Model:        fmt.Sprintf("Model-%d", i+1),
				Mass:         genRandomInt(MASS_MIN, MASS_MAX),
				MaxAltitude:  genRandomInt(ALTITUDE_MIN, ALTITUDE_MAX),
				MaxSpeed:     genRandomInt(SPEED_MIN, SPEED_MAX),
			})
		}
	}

	rows := make([][]any, 0, len(aircraftModels))
	for _, m := range aircraftModels {
		rows = append(rows, []any{
			m.ID,
			m.Manufacturer,
			m.Model,
			m.Mass,
			m.MaxAltitude,
			m.MaxSpeed,
		})
	}

	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"aircraft_models"},
		[]string{"id", "manufacturer", "model", "mass", "max_altitude", "max_speed"},
		pgx.CopyFromRows(rows),
	)

	return err
}

func GenAircrafts(ctx context.Context, conn *pgxpool.Pool) error {
	for i := 0; i < AIRCRAFTS_AMT; i++ {
		aircraftModel := randomChoice(aircraftModels)
		amt := genRandomInt(5, 10)

		for j := 0; j < amt; j++ {
			number := len(aircrafts) + 1

			aircrafts = append(aircrafts, Aircraft{
				ID:                 genUUID(),
				AircraftModelID:    aircraftModel.ID,
				RegistrationNumber: fmt.Sprintf("R%09d", number),
				SerialNumber:       fmt.Sprintf("S%09d", number),
				Mileage:            genRandomInt(MILEAGE_MIN, MILEAGE_MAX),
			})
		}
	}

	rows := make([][]any, 0, len(aircrafts))
	for _, a := range aircrafts {
		rows = append(rows, []any{
			a.ID,
			a.AircraftModelID,
			a.RegistrationNumber,
			a.SerialNumber,
			a.Mileage,
		})
	}

	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"aircraft"},
		[]string{"id", "aircraft_model_id", "registration_number", "serial_number", "mileage"},
		pgx.CopyFromRows(rows),
	)

	return err
}

func genFlightPlan() *string {
	length := genRandomInt(0, 200)

	baseChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	extraChars := "- "

	var b strings.Builder

	for i := 0; i < length; i++ {
		if rand.Float64() < 0.1 {
			b.WriteByte(extraChars[rand.Intn(len(extraChars))])
		} else {
			b.WriteByte(baseChars[rand.Intn(len(baseChars))])
		}
	}

	result := strings.TrimSpace(b.String())

	if len(result) < 2 {
		return nil
	}

	return &result
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func GenFlights(ctx context.Context, conn *pgxpool.Pool) error {
	now := time.Now().UTC()

	scenarios := []string{
		"future_scheduled",
		"future_boarding",
		"future_delayed",
		"future_cancelled",
		"future_rescheduled",
		"in_air",
		"landed_not_arrived",
		"completed",
		"past_cancelled",
	}

	weights := []int{25, 7, 10, 5, 5, 15, 8, 20, 5}

	for i := 0; i < FLIGHTS_AMT; i++ {
		aircraft := randomChoice(aircrafts)

		flightID := genUUID()
		aircraftID := aircraft.ID

		scenario := weightedChoice(scenarios, weights)

		var status string
		var schDep time.Time
		var schArr time.Time
		var actDep *time.Time
		var actArr *time.Time

		switch scenario {
		case "future_scheduled":
			status = "scheduled"
			schDep = now.Add(time.Duration(genRandomInt(2, 72)) * time.Hour)
			schArr = schDep.Add(time.Duration(genRandomInt(1, 10)) * time.Hour)

		case "future_boarding":
			status = "boarding"
			schDep = now.Add(time.Duration(genRandomInt(5, 45)) * time.Minute)
			schArr = schDep.Add(time.Duration(genRandomInt(1, 10)) * time.Hour)

		case "future_delayed":
			status = "delayed"
			schDep = now.Add(time.Duration(genRandomInt(30, 360)) * time.Minute)
			schArr = schDep.Add(time.Duration(genRandomInt(1, 10)) * time.Hour)

		case "future_cancelled":
			status = "cancelled"
			schDep = now.Add(-time.Duration(genRandomInt(1, 72)) * time.Hour)
			schArr = schDep.Add(time.Duration(genRandomInt(1, 10)) * time.Hour)

		case "future_rescheduled":
			status = "rescheduled"
			schDep = now.Add(time.Duration(genRandomInt(2, 96)) * time.Hour)
			schArr = schDep.Add(time.Duration(genRandomInt(1, 10)) * time.Hour)

		case "in_air":
			status = "departed"
			schDep = now.Add(-time.Duration(genRandomInt(20, 240)) * time.Minute)
			schArr = now.Add(time.Duration(genRandomInt(20, 360)) * time.Minute)

			tmpActDep := schDep.Add(time.Duration(genRandomInt(-10, 60)) * time.Minute)

			if tmpActDep.After(now) {
				tmpActDep = now.Add(-time.Duration(genRandomInt(5, 19)) * time.Minute)
			}

			actDep = ptrTime(tmpActDep)
			actArr = nil

		case "landed_not_arrived":
			status = "landed"
			schDep = now.Add(-time.Duration(genRandomInt(2, 12)) * time.Hour)
			schArr = now.Add(-time.Duration(genRandomInt(5, 60)) * time.Minute)

			tmpActDep := schDep.Add(time.Duration(genRandomInt(-10, 60)) * time.Minute)
			tmpActArr := now.Add(-time.Duration(genRandomInt(1, 30)) * time.Minute)

			actDep = ptrTime(tmpActDep)
			actArr = ptrTime(tmpActArr)

		case "completed":
			status = "arrived"
			schDep = now.Add(-time.Duration(genRandomInt(3, 72)) * time.Hour)
			schArr = schDep.Add(time.Duration(genRandomInt(1, 10)) * time.Hour)

			tmpActDep := schDep.Add(time.Duration(genRandomInt(-10, 60)) * time.Minute)
			tmpActArr := schArr.Add(time.Duration(genRandomInt(-20, 90)) * time.Minute)

			if !tmpActArr.After(tmpActDep) {
				tmpActArr = tmpActDep.Add(time.Duration(genRandomInt(30, 600)) * time.Minute)
			}

			if tmpActArr.After(now) {
				tmpActArr = now.Add(-time.Duration(genRandomInt(1, 60)) * time.Minute)
			}

			if !tmpActArr.After(tmpActDep) {
				tmpActDep = tmpActArr.Add(-time.Duration(genRandomInt(30, 600)) * time.Minute)
			}

			actDep = ptrTime(tmpActDep)
			actArr = ptrTime(tmpActArr)

		case "past_cancelled":
			status = "cancelled"
			schDep = now.Add(-time.Duration(genRandomInt(2, 72)) * time.Hour)
			schArr = schDep.Add(time.Duration(genRandomInt(1, 10)) * time.Hour)
		}

		flights = append(flights, Flight{
			ID:                 flightID,
			AircraftID:         aircraftID,
			ScheduledDeparture: schDep,
			ScheduledArrival:   schArr,
			ActualDeparture:    actDep,
			ActualArrival:      actArr,
			Status:             status,
			Plan:               genFlightPlan(),
		})
	}

	rows := make([][]any, 0, len(flights))
	for _, f := range flights {
		rows = append(rows, []any{
			f.ID,
			f.AircraftID,
			f.ScheduledDeparture,
			f.ScheduledArrival,
			f.ActualDeparture,
			f.ActualArrival,
			f.Status,
			f.Plan,
		})
	}

	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"flights"},
		[]string{
			"id",
			"aircraft_id",
			"scheduled_departure",
			"scheduled_arrival",
			"actual_departure",
			"actual_arrival",
			"status",
			"plan",
		},
		pgx.CopyFromRows(rows),
	)

	return err
}

func weightedChoice(items []string, weights []int) string {
	total := 0
	for _, w := range weights {
		total += w
	}

	r := rand.Intn(total)

	for i, w := range weights {
		if r < w {
			return items[i]
		}
		r -= w
	}

	return items[len(items)-1]
}

func GenFlightRoutes(ctx context.Context, conn *pgxpool.Pool) error {
	for _, flight := range flights {
		departureGate := randomChoice(gates)
		arrivalGate, err := gateFromDifferentAirport(departureGate)
		if err != nil {
			return err
		}

		flightRoutes = append(flightRoutes, FlightRoute{
			ID:              genUUID(),
			FlightID:        flight.ID,
			DepartureGateID: departureGate.ID,
			ArrivalGateID:   arrivalGate.ID,
		})
	}

	rows := make([][]any, 0, len(flightRoutes))
	for _, r := range flightRoutes {
		rows = append(rows, []any{
			r.ID,
			r.FlightID,
			r.DepartureGateID,
			r.ArrivalGateID,
		})
	}

	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"flight_routes"},
		[]string{"id", "flight_id", "departure_gate_id", "arrival_gate_id"},
		pgx.CopyFromRows(rows),
	)

	return err
}

func gateFromDifferentAirport(departureGate Gate) (Gate, error) {
	start := rand.Intn(len(gates))
	for offset := 0; offset < len(gates); offset++ {
		candidate := gates[(start+offset)%len(gates)]
		if gateToAirport[candidate.ID] != gateToAirport[departureGate.ID] {
			return candidate, nil
		}
	}

	return Gate{}, fmt.Errorf("arrival gate from a different airport not found")
}

func GenSubscriptions(ctx context.Context, conn *pgxpool.Pool) error {
	for userIndex, user := range users {
		for flightOffset := 0; flightOffset < min(3, len(flights)); flightOffset++ {
			flightIndex := (userIndex + flightOffset) % len(flights)
			subscriptions = append(subscriptions, Subscription{
				ID:       genUUID(),
				UserID:   user.ID,
				FlightID: flights[flightIndex].ID,
			})
		}
	}

	rows := make([][]any, 0, len(subscriptions))
	for _, s := range subscriptions {
		rows = append(rows, []any{
			s.ID,
			s.UserID,
			s.FlightID,
		})
	}

	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"subscriptions"},
		[]string{"id", "user_id", "flight_id"},
		pgx.CopyFromRows(rows),
	)

	return err
}

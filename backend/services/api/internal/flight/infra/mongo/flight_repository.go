package mongo

import (
	airportDomain "api/internal/airport/domain"
	airportMongoModel "api/internal/airport/infra/mongo/model"
	flightDomain "api/internal/flight/domain"
	"api/internal/flight/domain/repository"
	flightMongoModel "api/internal/flight/infra/mongo/model"
	userDomain "api/internal/user/domain"
	"context"
	"errors"
	"fmt"
	"shared/common"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDB struct {
	db *mongo.Database
}

func NewMongoDB(i do.Injector) (*MongoDB, error) {
	db := do.MustInvoke[*mongo.Database](i)
	return &MongoDB{
		db: db,
	}, nil
}

func (r *MongoDB) Save(ctx context.Context, f flightDomain.Flight) error {
	const op = "MongoDB.Save"

	doc := flightMongoModel.ToFlightDoc(f)
	collection := r.db.Collection("flights")

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, we := range writeException.WriteErrors {
				if we.Code == 11000 {
					return nil
				}
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *MongoDB) Exist(ctx context.Context, fid uuid.UUID) (flightDomain.Flight, error) {
	const op = "MongoDB.Exist"

	collection := r.db.Collection("flights")
	cursor, err := collection.Aggregate(ctx, flightInfoPipeline(bson.D{{Key: "_id", Value: fid.String()}}))
	if err != nil {
		return flightDomain.Flight{}, fmt.Errorf("%s: aggregate error: %w", op, err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		if err := cursor.Err(); err != nil {
			return flightDomain.Flight{}, fmt.Errorf("%s: cursor error: %w", op, err)
		}
		return flightDomain.Flight{}, repository.ErrFlightNotFound
	}

	var doc flightMongoModel.FlightDoc
	if err := cursor.Decode(&doc); err != nil {
		return flightDomain.Flight{}, fmt.Errorf("%s: decode error: %w", op, err)
	}

	fd, err := flightMongoModel.FromFlightDoc(doc)
	if err != nil {
		return flightDomain.Flight{}, fmt.Errorf("%s: conversion error: %w", op, err)
	}

	return fd, nil
}

func (r *MongoDB) Update(ctx context.Context, ufi flightDomain.UpdateFlightInfo) error {
	const op = "MongoDB.Update"

	collection := r.db.Collection("flights")
	filter := bson.M{"_id": ufi.FlightId.String()}

	updateFields := bson.M{}

	if ufi.ScheduledDeparture != nil {
		updateFields["scheduled_departure"] = *ufi.ScheduledDeparture
	}
	if ufi.ActualDeparture != nil {
		updateFields["actual_departure"] = *ufi.ActualDeparture
	}
	if ufi.ScheduledArrival != nil {
		updateFields["scheduled_arrival"] = *ufi.ScheduledArrival
	}
	if ufi.ActualArrival != nil {
		updateFields["actual_arrival"] = *ufi.ActualArrival
	}
	if ufi.Status != nil {
		updateFields["status"] = *ufi.Status
	}
	if ufi.Plan != nil {
		updateFields["plan"] = *ufi.Plan
	}

	if len(updateFields) == 0 {
		return nil
	}

	update := bson.M{"$set": updateFields}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func flightInfoPipeline(match bson.D) mongo.Pipeline {
	pipeline := mongo.Pipeline{}
	if len(match) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	return append(pipeline,
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "flight_routes"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "flight_id"},
			{Key: "as", Value: "route_info"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$route_info"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "departure_gate_id", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$departure_gate_id", "$route_info.departure_gate_id"}}}},
			{Key: "arrival_gate_id", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$arrival_gate_id", "$route_info.arrival_gate_id"}}}},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "gates"},
			{Key: "localField", Value: "departure_gate_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "departure_gate"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$departure_gate"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "gates"},
			{Key: "localField", Value: "arrival_gate_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "arrival_gate"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$arrival_gate"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "departure_airport_id", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$departure_airport_id", "$departure_gate.airport_id"}}}},
			{Key: "arrival_airport_id", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$arrival_airport_id", "$arrival_gate.airport_id"}}}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "route_info", Value: 0},
			{Key: "departure_gate", Value: 0},
			{Key: "arrival_gate", Value: 0},
		}}},
	)
}

func (r *MongoDB) ListFlights(ctx context.Context) ([]flightDomain.Flight, error) {
	const op = "MongoDB.ListFlights"

	collection := r.db.Collection("flights")
	cursor, err := collection.Aggregate(ctx, flightInfoPipeline(nil))
	if err != nil {
		return nil, fmt.Errorf("%s: aggregate error: %w", op, err)
	}
	defer cursor.Close(ctx)

	var docs []flightMongoModel.FlightDoc
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("%s: decode error: %w", op, err)
	}

	flights := make([]flightDomain.Flight, 0, len(docs))
	for _, doc := range docs {
		f, err := flightMongoModel.FromFlightDoc(doc)
		if err != nil {
			return nil, fmt.Errorf("%s: convert flight %s: %w", op, doc.ID, err)
		}
		flights = append(flights, f)
	}

	return flights, nil
}

func (r *MongoDB) GetFlightRoute(ctx context.Context, fid uuid.UUID) (flightDomain.FlightRoute, error) {
	const op = "MongoDB.GetFlightRoute"

	var routeDoc struct {
		ID              string `bson:"_id"`
		FlightID        string `bson:"flight_id"`
		DepartureGateID string `bson:"departure_gate_id"`
		ArrivalGateID   string `bson:"arrival_gate_id"`
	}

	err := r.db.Collection("flight_routes").FindOne(ctx, bson.M{"flight_id": fid.String()}).Decode(&routeDoc)
	if err == nil {
		routeID, err := uuid.Parse(routeDoc.ID)
		if err != nil {
			return flightDomain.FlightRoute{}, fmt.Errorf("%s: parse route id: %w", op, err)
		}
		flightID, err := uuid.Parse(routeDoc.FlightID)
		if err != nil {
			return flightDomain.FlightRoute{}, fmt.Errorf("%s: parse flight id: %w", op, err)
		}
		departureGateID, err := uuid.Parse(routeDoc.DepartureGateID)
		if err != nil {
			return flightDomain.FlightRoute{}, fmt.Errorf("%s: parse departure gate id: %w", op, err)
		}
		arrivalGateID, err := uuid.Parse(routeDoc.ArrivalGateID)
		if err != nil {
			return flightDomain.FlightRoute{}, fmt.Errorf("%s: parse arrival gate id: %w", op, err)
		}
		return flightDomain.FlightRoute{
			Id:              routeID,
			FlightId:        flightID,
			DepartureGateId: departureGateID,
			ArrivalGateId:   arrivalGateID,
		}, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return flightDomain.FlightRoute{}, fmt.Errorf("%s: %w", op, err)
	}

	flight, err := r.Exist(ctx, fid)
	if err != nil {
		return flightDomain.FlightRoute{}, err
	}

	route := flightDomain.FlightRoute{
		FlightId:        flight.Id,
		DepartureGateId: flight.DepartureGateId,
		ArrivalGateId:   flight.ArrivalGateId,
	}

	return route, nil
}

func (r *MongoDB) ListSubscribers(ctx context.Context, fid uuid.UUID) ([]userDomain.User, error) {
	const op = "MongoDB.ListSubscribers"

	subscriptionsCol := r.db.Collection("subscriptions")
	usersCol := r.db.Collection("users")

	filter := bson.M{"flight_id": fid.String()}
	cursor, err := subscriptionsCol.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%s: find subscriptions: %w", op, err)
	}

	type SubDoc struct {
		UserID string `bson:"user_id"`
	}
	var subs []SubDoc
	if err = cursor.All(ctx, &subs); err != nil {
		return nil, fmt.Errorf("%s: decode subscriptions: %w", op, err)
	}

	if len(subs) == 0 {
		return []userDomain.User{}, nil
	}

	userIDs := make([]string, len(subs))
	for i, sub := range subs {
		userIDs[i] = sub.UserID
	}

	usersFilter := bson.M{"_id": bson.M{"$in": userIDs}}
	usersCursor, err := usersCol.Find(ctx, usersFilter)
	if err != nil {
		return nil, fmt.Errorf("%s: find users: %w", op, err)
	}
	defer usersCursor.Close(ctx)

	type UserDoc struct {
		ID           string `bson:"_id"`
		Email        string `bson:"email"`
		PasswordHash string `bson:"password_hash"`
		Role         string `bson:"role"`
	}

	var userDocs []UserDoc
	if err = usersCursor.All(ctx, &userDocs); err != nil {
		return nil, fmt.Errorf("%s: decode users: %w", op, err)
	}

	users := make([]userDomain.User, 0, len(userDocs))
	for _, uDoc := range userDocs {
		u, err := convertUserDocToDomain(uDoc)
		if err != nil {
			continue
		}
		users = append(users, u)
	}

	return users, nil
}

func convertUserDocToDomain(doc struct {
	ID           string `bson:"_id"`
	Email        string `bson:"email"`
	PasswordHash string `bson:"password_hash"`
	Role         string `bson:"role"`
}) (userDomain.User, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return userDomain.User{}, err
	}
	email, err := common.NewEmail(doc.Email)
	if err != nil {
		return userDomain.User{}, err
	}

	pass, err := userDomain.NewPasswordHashed(doc.PasswordHash)
	if err != nil {
		return userDomain.User{}, err
	}
	role, err := userDomain.NewRole(doc.Role)
	if err != nil {
		return userDomain.User{}, err
	}

	return userDomain.User{
		Id:           id,
		Email:        email,
		PasswordHash: pass,
		Role:         role,
	}, nil
}

func (r *MongoDB) GetFlightAirports(
	ctx context.Context,
	fid uuid.UUID,
) (dep airportDomain.Airport, arr airportDomain.Airport, err error) {
	const op = "MongoDB.GetFlightAirports"

	flightsCol := r.db.Collection("flights")

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"_id": fid.String()}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "flight_routes",
			"localField":   "_id",
			"foreignField": "flight_id",
			"as":           "route_info",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$route_info",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$addFields", Value: bson.M{
			"departure_gate_id": bson.M{"$ifNull": bson.A{"$departure_gate_id", "$route_info.departure_gate_id"}},
			"arrival_gate_id":   bson.M{"$ifNull": bson.A{"$arrival_gate_id", "$route_info.arrival_gate_id"}},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "gates",
			"localField":   "departure_gate_id",
			"foreignField": "_id",
			"as":           "dep_gate",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$dep_gate"}}},

		{{Key: "$lookup", Value: bson.M{
			"from":         "gates",
			"localField":   "arrival_gate_id",
			"foreignField": "_id",
			"as":           "arr_gate",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$arr_gate"}}},

		{{Key: "$lookup", Value: bson.M{
			"from":         "airports",
			"localField":   "dep_gate.airport_id",
			"foreignField": "_id",
			"as":           "dep_airport",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$dep_airport"}}},

		{{Key: "$lookup", Value: bson.M{
			"from":         "airports",
			"localField":   "arr_gate.airport_id",
			"foreignField": "_id",
			"as":           "arr_airport",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$arr_airport"}}},

		{{Key: "$project", Value: bson.M{
			"dep_airport": 1,
			"arr_airport": 1,
			"_id":         0,
		}}},
	}

	cursor, err := flightsCol.Aggregate(ctx, pipeline)
	if err != nil {
		return dep, arr, fmt.Errorf("%s: aggregate error: %w", op, err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return dep, arr, repository.ErrFlightNotFound
	}

	var result struct {
		DepAirport airportMongoModel.AirportModel `bson:"dep_airport"`
		ArrAirport airportMongoModel.AirportModel `bson:"arr_airport"`
	}

	if err = cursor.Decode(&result); err != nil {
		return dep, arr, fmt.Errorf("%s: decode error: %w", op, err)
	}

	dep, err = airportMongoModel.ModelToAirport(result.DepAirport)
	if err != nil {
		return dep, arr, fmt.Errorf("%s: convert dep airport: %w", op, err)
	}

	arr, err = airportMongoModel.ModelToAirport(result.ArrAirport)
	if err != nil {
		return dep, arr, fmt.Errorf("%s: convert arr airport: %w", op, err)
	}

	return dep, arr, nil
}

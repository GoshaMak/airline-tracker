package mongo

import (
	flightDomain "api/internal/flight/domain"
	flightModel "api/internal/flight/infra/mongo/model"
	"api/internal/user/domain"
	"api/internal/user/domain/repository"
	"api/internal/user/infra/mongo/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userRepository struct {
	db *mongo.Database
}

func NewUserRepository(i do.Injector) (repository.UserRepository, error) {
	db := do.MustInvoke[*mongo.Database](i)

	usersCol := db.Collection("users")
	subsCol := db.Collection("subscriptions")

	usersIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	usersCol.Indexes().CreateOne(context.Background(), usersIndex)

	subsIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "flight_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	subsCol.Indexes().CreateOne(context.Background(), subsIndex)

	return &userRepository{
		db: db,
	}, nil
}

func (r *userRepository) SaveUser(ctx context.Context, user domain.User) error {
	const op = "UserRepository.SaveUser"

	doc := model.ToUserDoc(user)
	collection := r.db.Collection("users")

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, we := range writeException.WriteErrors {
				if we.Code == 11000 {
					return repository.ErrUserAlreadyExists
				}
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *userRepository) GetUser(ctx context.Context, email string) (domain.User, error) {
	const op = "UserRepository.GetUser"

	collection := r.db.Collection("users")
	filter := bson.M{"email": email}

	var doc model.UserDoc
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	ud, err := model.FromUserDoc(doc)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: conversion error: %w", op, err)
	}

	return ud, nil
}

func (r *userRepository) Exist(ctx context.Context, uid uuid.UUID) (domain.User, error) {
	const op = "UserRepository.Exist"

	collection := r.db.Collection("users")
	filter := bson.M{"_id": uid.String()}

	var doc model.UserDoc
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	ud, err := model.FromUserDoc(doc)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: conversion error: %w", op, err)
	}

	return ud, nil
}

func (r *userRepository) Subscribe(ctx context.Context, uid, fid uuid.UUID) error {
	const op = "UserRepository.Subscribe"

	usersCol := r.db.Collection("users")
	flightsCol := r.db.Collection("flights")
	subsCol := r.db.Collection("subscriptions")

	userFilter := bson.M{"_id": uid.String()}
	count, err := usersCol.CountDocuments(ctx, userFilter)
	if err != nil {
		return fmt.Errorf("%s: check user: %w", op, err)
	}
	if count == 0 {
		return repository.ErrUserNotFound
	}

	flightFilter := bson.M{"_id": fid.String()}
	count, err = flightsCol.CountDocuments(ctx, flightFilter)
	if err != nil {
		return fmt.Errorf("%s: check flight: %w", op, err)
	}
	if count == 0 {
		return repository.ErrFlightNotFound
	}

	subDoc := model.SubscriptionDoc{
		UserID:   uid.String(),
		FlightID: fid.String(),
	}

	_, err = subsCol.InsertOne(ctx, subDoc)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, we := range writeException.WriteErrors {
				if we.Code == 11000 {
					return repository.ErrUserAlreadySubscribed
				}
			}
		}
		return fmt.Errorf("%s: insert subscription: %w", op, err)
	}

	return nil
}

func (r *userRepository) ListFlights(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error) {
	const op = "UserRepository.ListFlights"

	subsCol := r.db.Collection("subscriptions")
	flightsCol := r.db.Collection("flights")

	subsFilter := bson.M{"user_id": uid.String()}
	cursor, err := subsCol.Find(ctx, subsFilter)
	if err != nil {
		return nil, fmt.Errorf("%s: find subscriptions: %w", op, err)
	}

	var subs []model.SubscriptionDoc
	if err = cursor.All(ctx, &subs); err != nil {
		return nil, fmt.Errorf("%s: decode subscriptions: %w", op, err)
	}

	if len(subs) == 0 {
		return []flightDomain.Flight{}, nil
	}

	flightIDs := make([]string, len(subs))
	for i, sub := range subs {
		flightIDs[i] = sub.FlightID
	}

	flightsCursor, err := flightsCol.Aggregate(ctx, userFlightsInfoPipeline(flightIDs))
	if err != nil {
		return nil, fmt.Errorf("%s: find flights: %w", op, err)
	}
	defer flightsCursor.Close(ctx)

	var flightDocs []flightModel.FlightDoc
	if err = flightsCursor.All(ctx, &flightDocs); err != nil {
		return nil, fmt.Errorf("%s: decode flights: %w", op, err)
	}

	flights := make([]flightDomain.Flight, 0, len(flightDocs))
	for _, doc := range flightDocs {
		f, err := flightModel.FromFlightDoc(doc)
		if err != nil {
			return nil, fmt.Errorf("%s: convert flight %s: %w", op, doc.ID, err)
		}
		flights = append(flights, f)
	}

	return flights, nil
}

func userFlightsInfoPipeline(flightIDs []string) mongo.Pipeline {
	return mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: flightIDs}}}}}},
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
	}
}

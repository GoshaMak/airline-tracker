package mongo

import (
	"api/internal/fleet/domain"
	"api/internal/fleet/domain/repository"
	"api/internal/fleet/infra/mongo/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type aircraftModelRepository struct {
	collection *mongo.Collection
}

func NewAircraftModelRepository(i do.Injector) (repository.AircraftModelRepository, error) {
	db := do.MustInvoke[*mongo.Database](i)
	return &aircraftModelRepository{
		collection: db.Collection("aircraft_models"),
	}, nil
}

func (r *aircraftModelRepository) SaveAircraftModel(ctx context.Context, am domain.AircraftModel) error {
	const op = "AircraftModelRepository.SaveAircraftModel"

	doc := model.ToAircraftModelDoc(am)

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, we := range writeException.WriteErrors {
				if we.Code == 11000 {
					return repository.ErrAircraftModelAlreadyExists
				}
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *aircraftModelRepository) GetAircraftModelById(ctx context.Context, id uuid.UUID) (domain.AircraftModel, error) {
	const op = "AircraftModelRepository.GetAircraftModelById"

	filter := bson.M{"_id": id.String()}

	var doc model.AircraftModelDoc
	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.AircraftModel{}, repository.ErrAircraftModelNotFound
		}
		return domain.AircraftModel{}, fmt.Errorf("%s: %w", op, err)
	}

	amd, err := model.FromAircraftModelDoc(doc)
	if err != nil {
		return domain.AircraftModel{}, fmt.Errorf("%s: conversion error: %w", op, err)
	}

	return amd, nil
}

package mongo

import (
	"api/internal/fleet/domain"
	"api/internal/fleet/domain/repository"
	"api/internal/fleet/infra/mongo/model"
	"context"
	"errors"
	"fmt"

	"github.com/samber/do/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type aircraftRepository struct {
	collection *mongo.Collection
}

func NewAircraftRepository(i do.Injector) (repository.AircraftRepository, error) {
	db := do.MustInvoke[*mongo.Database](i)
	return &aircraftRepository{
		collection: db.Collection("aircraft"),
	}, nil
}

func (r *aircraftRepository) SaveAircraft(ctx context.Context, a domain.Aircraft) error {
	const op = "AircraftRepository.SaveAircraft"

	doc := model.ToAircraftDoc(a)

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, we := range writeException.WriteErrors {
				if we.Code == 11000 {
					return repository.ErrAircraftAlreadyExists
				}
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *aircraftRepository) List(ctx context.Context) ([]domain.Aircraft, error) {
	const op = "AircraftRepository.List"

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer cursor.Close(ctx)

	var docs []model.AircraftDoc
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	aircrafts := make([]domain.Aircraft, 0, len(docs))
	for _, doc := range docs {
		a, err := model.FromAircraftDoc(doc)
		if err != nil {
			continue
		}
		aircrafts = append(aircrafts, a)
	}

	return aircrafts, nil
}

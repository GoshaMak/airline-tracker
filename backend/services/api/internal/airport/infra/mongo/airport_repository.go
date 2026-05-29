package mongo

import (
	"api/internal/airport/domain"
	"api/internal/airport/domain/repository"
	"api/internal/airport/infra/mongo/model"
	"context"
	"errors"
	"fmt"

	"github.com/samber/do/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type airportRepository struct {
	collection *mongo.Collection
}

func NewAirportRepository(i do.Injector) (repository.AirportRepository, error) {
	db := do.MustInvoke[*mongo.Database](i)
	return &airportRepository{
		collection: db.Collection("airports"),
	}, nil
}

func (r *airportRepository) Save(ctx context.Context, a domain.Airport) error {
	const op = "AirportRepository.Save"

	model := model.AirportToModel(a)

	_, err := r.collection.InsertOne(ctx, model)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, we := range writeException.WriteErrors {
				if we.Code == 11000 {
					return repository.ErrAirportAlreadyExists
				}
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *airportRepository) ListAirports(ctx context.Context) ([]domain.Airport, error) {
	const op = "AirportRepository.ListAirports"

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer cursor.Close(ctx)

	var models []model.AirportModel
	if err = cursor.All(ctx, &models); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(models) == 0 {
		return nil, nil
	}

	airports := make([]domain.Airport, 0, len(models))
	for _, m := range models {
		a, err := model.ModelToAirport(m)
		if err != nil {
			continue
		}
		airports = append(airports, a)
	}

	return airports, nil
}

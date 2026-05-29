package mongo

import (
	"api/internal/airport/domain"
	"api/internal/airport/domain/repository"
	"api/internal/airport/infra/mongo/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type gateRepositoryFixed struct {
	db *mongo.Database
}

func NewGateRepository(i do.Injector) (repository.GateRepository, error) {
	db := do.MustInvoke[*mongo.Database](i)
	return &gateRepositoryFixed{
		db: db,
	}, nil
}

func (r *gateRepositoryFixed) Save(ctx context.Context, g domain.Gate) error {
	const op = "GateRepository.Save"
	model := model.GateToModel(g)

	col := r.db.Collection("gates")
	_, err := col.InsertOne(ctx, model)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, we := range writeException.WriteErrors {
				if we.Code == 11000 {
					return repository.ErrGateAlreadyExists
				}
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *gateRepositoryFixed) GetAirportByGateId(ctx context.Context, gid uuid.UUID) (domain.Airport, error) {
	const op = "GateRepository.GetAirportByGateId"

	gatesCol := r.db.Collection("gates")
	airportsCol := r.db.Collection("airports")

	var gateModel model.GateModel
	err := gatesCol.FindOne(ctx, bson.M{"_id": gid.String()}).Decode(&gateModel)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Airport{}, repository.ErrAirportNotFound
		}
		return domain.Airport{}, fmt.Errorf("%s: find gate: %w", op, err)
	}

	var airportModel model.AirportModel
	err = airportsCol.FindOne(ctx, bson.M{"_id": gateModel.AirportID}).Decode(&airportModel)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Airport{}, repository.ErrAirportNotFound
		}
		return domain.Airport{}, fmt.Errorf("%s: find airport: %w", op, err)
	}

	ad, err := model.ModelToAirport(airportModel)
	if err != nil {
		return domain.Airport{}, fmt.Errorf("%s: convert model: %w", op, err)
	}

	return ad, nil
}

func (r *gateRepositoryFixed) List(ctx context.Context) ([]domain.Gate, error) {
	const op = "GateRepository.List"

	col := r.db.Collection("gates")
	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer cursor.Close(ctx)

	var models []model.GateModel
	if err = cursor.All(ctx, &models); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	gates := make([]domain.Gate, 0, len(models))
	for _, m := range models {
		g, err := model.ModelToGate(m)
		if err != nil {
			continue
		}
		gates = append(gates, g)
	}

	return gates, nil
}

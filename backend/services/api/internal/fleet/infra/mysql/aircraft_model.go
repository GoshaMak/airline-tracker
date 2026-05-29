package mysql

import (
	"api/internal/fleet/domain"
	"api/internal/fleet/domain/repository"
	dbmysql "api/internal/infra/mysql"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type aircraftModelRepository struct {
	conn *sql.DB
}

func NewAircraftModelRepository(i do.Injector) (repository.AircraftModelRepository, error) {
	return &aircraftModelRepository{
		conn: do.MustInvoke[*sql.DB](i),
	}, nil
}

func (r *aircraftModelRepository) SaveAircraftModel(
	ctx context.Context,
	am domain.AircraftModel,
) error {
	const op = "AircraftModelRepository.SaveAircraftModel"
	query := `
	insert into
		aircraft_models(manufacturer, model, mass, max_altitude, max_speed)
		values (?, ?, ?, ?, ?)
	`
	_, err := r.conn.ExecContext(ctx, query,
		am.Manufacturer.String(), am.Model.String(),
		am.Mass.Value(), am.MaxAltitude.Value(), am.MaxSpeed.Value(),
	)
	if err != nil {
		if dbmysql.IsDuplicate(err) {
			return repository.ErrAircraftModelAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *aircraftModelRepository) GetAircraftModelById(
	ctx context.Context,
	id uuid.UUID,
) (domain.AircraftModel, error) {
	const op = "AircraftModelRepository.GetAircraftModelById"
	query := `
	select id, manufacturer, model, mass, max_altitude, max_speed
	from aircraft_models
	where id = ?
	`
	am, err := scanAircraftModel(r.conn.QueryRowContext(ctx, query, id.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.AircraftModel{}, repository.ErrAircraftModelNotFound
		}
		return domain.AircraftModel{}, fmt.Errorf("%s: %w", op, err)
	}

	return am, nil
}

func scanAircraftModel(scanner fleetScanner) (domain.AircraftModel, error) {
	var (
		id                         uuid.UUID
		manufacturer, model        string
		mass, maxAltitude, maxSped int
	)
	if err := scanner.Scan(&id, &manufacturer, &model, &mass, &maxAltitude, &maxSped); err != nil {
		return domain.AircraftModel{}, err
	}

	man, err := domain.NewManufacturer(manufacturer)
	if err != nil {
		return domain.AircraftModel{}, err
	}
	mod, err := domain.NewModel(model)
	if err != nil {
		return domain.AircraftModel{}, err
	}
	massValue, err := domain.NewAircraftMass(mass)
	if err != nil {
		return domain.AircraftModel{}, err
	}
	alt, err := domain.NewAircraftMaxAltitude(maxAltitude)
	if err != nil {
		return domain.AircraftModel{}, err
	}
	spd, err := domain.NewAircraftMaxSpeed(maxSped)
	if err != nil {
		return domain.AircraftModel{}, err
	}

	return domain.AircraftModel{
		Id:           id,
		Manufacturer: man,
		Model:        mod,
		Mass:         massValue,
		MaxAltitude:  alt,
		MaxSpeed:     spd,
	}, nil
}

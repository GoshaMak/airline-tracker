package mysql

import (
	"api/internal/fleet/domain"
	"api/internal/fleet/domain/repository"
	dbmysql "api/internal/infra/mysql"
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type aircraftRepository struct {
	conn *sql.DB
}

func NewAircraftRepository(i do.Injector) (repository.AircraftRepository, error) {
	return &aircraftRepository{
		conn: do.MustInvoke[*sql.DB](i),
	}, nil
}

func (r *aircraftRepository) SaveAircraft(
	ctx context.Context,
	a domain.Aircraft,
) error {
	const op = "AircraftRepository.SaveAircraft"
	query := `
	insert into
		aircraft(registration_number, aircraft_model_id, serial_number, mileage)
		values (?, ?, ?, ?)
	`
	_, err := r.conn.ExecContext(ctx, query,
		a.RegistrationNumber.String(), a.AircraftModelId.String(),
		a.SerialNumber.String(), a.Mileage.Value(),
	)
	if err != nil {
		if dbmysql.IsDuplicate(err) {
			return repository.ErrAircraftAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *aircraftRepository) List(ctx context.Context) ([]domain.Aircraft, error) {
	const op = "AircraftRepository.List"
	query := `
	select id, aircraft_model_id, registration_number, serial_number, mileage
	from aircraft
	`
	rows, err := r.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	aircrafts := make([]domain.Aircraft, 0)
	for rows.Next() {
		a, err := scanAircraft(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		aircrafts = append(aircrafts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return aircrafts, nil
}

type fleetScanner interface {
	Scan(dest ...any) error
}

func scanAircraft(scanner fleetScanner) (domain.Aircraft, error) {
	var (
		id, aircraftModelId           uuid.UUID
		registrationNumber, serialNum string
		mileage                       int
	)
	if err := scanner.Scan(&id, &aircraftModelId, &registrationNumber, &serialNum, &mileage); err != nil {
		return domain.Aircraft{}, err
	}

	rn, err := domain.NewRegistrationNumber(registrationNumber)
	if err != nil {
		return domain.Aircraft{}, err
	}
	sn, err := domain.NewSerialNumber(serialNum)
	if err != nil {
		return domain.Aircraft{}, err
	}
	m, err := domain.NewMileage(mileage)
	if err != nil {
		return domain.Aircraft{}, err
	}

	return domain.Aircraft{
		Id:                 id,
		AircraftModelId:    aircraftModelId,
		RegistrationNumber: rn,
		SerialNumber:       sn,
		Mileage:            m,
	}, nil
}

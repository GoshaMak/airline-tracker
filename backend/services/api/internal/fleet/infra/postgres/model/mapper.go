package model

import "api/internal/fleet/domain"

func AircraftModelToDomain(am AircraftModel) (domain.Aircraft, error) {
	rn, err := domain.NewRegistrationNumber(am.RegistrationNumber)
	if err != nil {
		return domain.Aircraft{}, err
	}
	sn, err := domain.NewSerialNumber(am.SerialNumber)
	if err != nil {
		return domain.Aircraft{}, err
	}
	m, err := domain.NewMileage(am.Mileage)
	if err != nil {
		return domain.Aircraft{}, err
	}
	return domain.Aircraft{
		Id:                 am.Id,
		AircraftModelId:    am.AircraftModelId,
		RegistrationNumber: rn,
		SerialNumber:       sn,
		Mileage:            m,
	}, nil
}

func AircraftModelModelToDomain(amm AircraftModelModel) (domain.AircraftModel, error) {
	man, err := domain.NewManufacturer(amm.Manufacturer)
	if err != nil {
		return domain.AircraftModel{}, err
	}
	mod, err := domain.NewModel(amm.Model)
	if err != nil {
		return domain.AircraftModel{}, err
	}
	mass, err := domain.NewAircraftMass(amm.Mass)
	if err != nil {
		return domain.AircraftModel{}, err
	}
	alt, err := domain.NewAircraftMaxAltitude(amm.MaxAltitude)
	if err != nil {
		return domain.AircraftModel{}, err
	}
	spd, err := domain.NewAircraftMaxSpeed(amm.MaxSpeed)
	if err != nil {
		return domain.AircraftModel{}, err
	}
	return domain.AircraftModel{
		Id:           amm.Id,
		Manufacturer: man,
		Model:        mod,
		Mass:         mass,
		MaxAltitude:  alt,
		MaxSpeed:     spd,
	}, nil
}

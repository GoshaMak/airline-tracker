package model

import (
	"api/internal/fleet/domain"

	"github.com/google/uuid"
)

type AircraftModelDoc struct {
	ID           string `bson:"_id"`
	Manufacturer string `bson:"manufacturer"`
	Model        string `bson:"model"`
	Mass         int    `bson:"mass"`
	MaxAltitude  int    `bson:"max_altitude"`
	MaxSpeed     int    `bson:"max_speed"`
}

func ToAircraftModelDoc(am domain.AircraftModel) AircraftModelDoc {
	return AircraftModelDoc{
		ID:           am.Id.String(),
		Manufacturer: am.Manufacturer.String(),
		Model:        am.Model.String(),
		Mass:         am.Mass.Value(),
		MaxAltitude:  am.MaxAltitude.Value(),
		MaxSpeed:     am.MaxSpeed.Value(),
	}
}

func FromAircraftModelDoc(doc AircraftModelDoc) (domain.AircraftModel, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return domain.AircraftModel{}, err
	}

	manufacturer, err := domain.NewManufacturer(doc.Manufacturer)
	if err != nil {
		return domain.AircraftModel{}, err
	}

	model, err := domain.NewModel(doc.Model)
	if err != nil {
		return domain.AircraftModel{}, err
	}

	mass, err := domain.NewAircraftMass(doc.Mass)
	if err != nil {
		return domain.AircraftModel{}, err
	}

	maxAlt, err := domain.NewAircraftMaxAltitude(doc.MaxAltitude)
	if err != nil {
		return domain.AircraftModel{}, err
	}

	maxSpd, err := domain.NewAircraftMaxSpeed(doc.MaxSpeed)
	if err != nil {
		return domain.AircraftModel{}, err
	}

	return domain.AircraftModel{
		Id:           id,
		Manufacturer: manufacturer,
		Model:        model,
		Mass:         mass,
		MaxAltitude:  maxAlt,
		MaxSpeed:     maxSpd,
	}, nil
}

type AircraftDoc struct {
	ID                 string `bson:"_id"`
	RegistrationNumber string `bson:"registration_number"`
	AircraftModelId    string `bson:"aircraft_model_id"`
	SerialNumber       string `bson:"serial_number"`
	Mileage            int    `bson:"mileage"`
}

func ToAircraftDoc(a domain.Aircraft) AircraftDoc {
	return AircraftDoc{
		ID:                 a.Id.String(),
		RegistrationNumber: a.RegistrationNumber.String(),
		AircraftModelId:    a.AircraftModelId.String(),
		SerialNumber:       a.SerialNumber.String(),
		Mileage:            a.Mileage.Value(),
	}
}

func FromAircraftDoc(doc AircraftDoc) (domain.Aircraft, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return domain.Aircraft{}, err
	}

	modelId, err := uuid.Parse(doc.AircraftModelId)
	if err != nil {
		return domain.Aircraft{}, err
	}

	regNum, err := domain.NewRegistrationNumber(doc.RegistrationNumber)
	if err != nil {
		return domain.Aircraft{}, err
	}

	serialNum, err := domain.NewSerialNumber(doc.SerialNumber)
	if err != nil {
		return domain.Aircraft{}, err
	}

	mileage, err := domain.NewMileage(doc.Mileage)
	if err != nil {
		return domain.Aircraft{}, err
	}

	return domain.Aircraft{
		Id:                 id,
		RegistrationNumber: regNum,
		AircraftModelId:    modelId,
		SerialNumber:       serialNum,
		Mileage:            mileage,
	}, nil
}

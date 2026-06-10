package domain

import (
	"errors"
	"strconv"
)

type Mileage int // miles

var (
	ErrNegativeMileage = errors.New("negative mileage value")
	ErrInvalidMileage  = errors.New("invalid mileage valued")
)

const (
	maxMileage = 50_000
)

func IsValidMileage(m int) error {
	if m < 0 {
		return ErrNegativeMileage
	}
	if m > maxMileage {
		return ErrInvalidMileage
	}
	return nil
}

func NewMileage(m int) (Mileage, error) {
	if err := IsValidMileage(m); err != nil {
		return -1, err
	}
	return Mileage(m), nil
}

func (m Mileage) String() string {
	return strconv.Itoa(int(m))
}

func (m Mileage) Value() int {
	return int(m)
}

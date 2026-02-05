package domain

import "time"

type Passport struct {
	ID           uint32      `json:"id"`
	Number       string      `json:"number"`
	IssueDate    time.Time   `json:"issue_date"`
	Name         string      `json:"name"`
	SecondName   string      `json:"second_name"`
	Surname      string      `json:"surname"`
	Gender       GenderState `json:"gender"`
	Birthday     time.Time   `json:"birthday"`
	BirthCity    string      `json:"birth_city"`
	BirthCountry string      `json:"birth_country"`
}

type GenderState int

const (
	Male GenderState = iota
	Female
)

var genderStateNames = map[GenderState]string{
	Male:   "male",
	Female: "female",
}

func (gs GenderState) String() string {
	return genderStateNames[gs]
}

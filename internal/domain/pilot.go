package domain

import "time"

type Pilot struct {
	ID         uint32     `json:"id"`
	Name       string     `json:"name"`
	Surname    string     `json:"surname"`
	Passport   Passport   `json:"passport"`
	Experience time.Time  `json:"experience"`
	Grade      PilotGrade `json:"grade"`
}

type PilotGrade int

const (
	Trainee PilotGrade = iota
	Junior
	FirstOfficer
	Captain
)

var pilotGradeNames = map[PilotGrade]string{
	Trainee:      "Trainee",
	Junior:       "Junior",
	FirstOfficer: "First Officer",
	Captain:      "Captain",
}

func (pg PilotGrade) String() string {
	return pilotGradeNames[pg]
}

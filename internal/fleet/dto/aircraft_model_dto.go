package dto

type AircraftModelDTO struct {
	Manufacturer string `json:"manufacturer" example:"A"`
	Model        string `json:"model" example:"B"`
	Mass         int    `json:"mass" example:"999"`
	MaxAltitude  int    `json:"max_altitude" example:"3"`
	MaxSpeed     int    `json:"max_speed" example:"1"`
}

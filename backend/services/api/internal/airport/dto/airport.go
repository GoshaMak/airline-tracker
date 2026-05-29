package dto

type AirportDTO struct {
	IATACode string `json:"iata_code" example:"SVO"`
	Title    string `json:"title" example:"123"`
	City     string `json:"city" example:"123"`
	Country  string `json:"country" example:"123"`
}

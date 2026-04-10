package dto

type ListFlightsRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

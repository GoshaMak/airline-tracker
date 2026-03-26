package dto

type ListFlightsRequest struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
}

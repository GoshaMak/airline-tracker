package dto

type LoginRequestDTO struct {
	Email    string `json:"email" example:"ab@cd.ef"`
	Password string `json:"password" example:"123"`
}

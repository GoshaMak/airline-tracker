package dto

type CreateUserDTO struct {
	Email    string `json:"email" example:"ab@cd.ef"`
	Password string `json:"password" example:"123"`
	Role     string `json:"role" example:"admin"`
}

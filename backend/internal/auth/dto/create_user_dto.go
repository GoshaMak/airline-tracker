package dto

type CreateUserDTO struct {
	Email    string `json:"email" example:"ab@cd.ef"`
	Password string `json:"password" example:"myStrong123"`
	Role     string `json:"role" example:"user"`
}

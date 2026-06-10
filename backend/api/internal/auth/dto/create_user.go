package dto

type CreateUserDTO struct {
	Email    string `json:"email" example:"user@user.user"`
	Password string `json:"password" example:"myStrong123!"`
}

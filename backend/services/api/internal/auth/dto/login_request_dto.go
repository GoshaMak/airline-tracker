package dto

type LoginRequestDTO struct {
	Email    string `json:"email" example:"a@b.c"`
	Password string `json:"password" example:"myStrong123"`
}

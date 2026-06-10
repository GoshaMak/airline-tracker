package dto

type LoginRequestDTO struct {
	Email    string `json:"email" example:"a@b.c/user=ab@cd.ef"`
	Password string `json:"password" example:"myStrong123"`
}

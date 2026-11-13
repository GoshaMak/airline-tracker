package dto

type LoginRequestDTO struct {
	Email    string `json:"email" binding:"required" example:"a@b.c/user=ab@cd.ef"`
	Password string `json:"password" binding:"required" example:"myStrong123"`
}

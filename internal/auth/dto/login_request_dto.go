package dto

type LoginRequestDTO struct {
	Email    string `json:"email" example:"ab@ab.de"`
	Phone    string `json:"phone" example:"01231231212"`
	Password string `json:"password" example:"123"`
	Role     string `json:"role" example:"admin"`
}

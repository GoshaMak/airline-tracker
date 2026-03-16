package dto

import "airline-tracker/internal/domain"

type UserDTO struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (u *UserDTO) UserFromDTO() *domain.User {
	return domain.NewUser(u.Email, u.Phone, u.Password, u.Role)
}

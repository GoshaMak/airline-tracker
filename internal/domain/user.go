package domain

type User struct {
	ID       uint
	Email    string
	Phone    string
	Password string
	Role     string
}

func NewUser(email, phone, password, role string) *User {
	return &User{
		Email:    email,
		Phone:    phone,
		Password: password,
		Role:     role,
	}
}

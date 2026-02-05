package domain

type User struct {
	ID         uint32 `json:"id"`
	PassportID uint32 `json:"passport_id"`
	CardID     uint32 `json:"card_id"`
	Login      string `json:"login"`
	Password   string `json:"password"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
}

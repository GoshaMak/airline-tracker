package domain

type User struct {
	ID         uint32 `json:"id"`
	PassportID uint32 `json:"passport_id,omitempty"`
	CardID     uint32 `json:"card_id,omitempty"`
	Password   string `json:"password"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
}

package repositories

type UserRepository interface {
	Save(u *int) error
}

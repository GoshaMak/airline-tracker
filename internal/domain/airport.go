package domain

type Airport struct {
	ID       uint32 `json:"id"`
	Title    string `json:"title"`
	Location string `json:"location"`
}

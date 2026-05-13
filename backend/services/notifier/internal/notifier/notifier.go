package notifier

type Notifier struct {
}

func NewNotifier() (*Notifier, error) {
	return &Notifier{}, nil
}

func (a *Notifier) Run() int {
	return 0
}

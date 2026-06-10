package sender

import (
	"context"
	"log/slog"
	"notifier/internal/sender/usecase"
	"time"

	"github.com/samber/do/v2"
)

type Sender struct {
	uc *usecase.SenderUsecase
}

func NewSender(i do.Injector) (*Sender, error) {
	return &Sender{
		uc: do.MustInvoke[*usecase.SenderUsecase](i),
	}, nil
}

func (s *Sender) Run(ctx context.Context) error {
	const op = "Sender.Run"
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			if err := s.uc.Send(ctx); err != nil {
				slog.Error(op, "err", err)
			}
		}
	}
}

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
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		processed, err := s.uc.Send(ctx)
		if err != nil {
			slog.ErrorContext(ctx, op, "err", err)
		}
		if processed && err == nil {
			continue
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

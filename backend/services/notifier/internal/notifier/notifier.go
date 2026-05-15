package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"notifier/internal/infra/kafka"
	"notifier/internal/notifier/handler"
)

type Notifier struct {
	nr *kafka.NotifyReceiver
	h  *handler.NotifierHandler
}

func NewNotifier() (*Notifier, error) {
	const op = "NewNotifier"
	nr, err := kafka.NewNotifyReceiver("flights")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	h, err := handler.NewNotifierHandler(nr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Notifier{
		nr: nr,
		h:  h,
	}, nil
}

func (n *Notifier) Run() error {
	const op = "Notifier.Run"
	go func() {
		if err := n.nr.StartConsuming(context.Background()); err != nil {
			slog.Error("error while consuming", "err", err)
		}
	}()

	for {
		msg, err := n.nr.ReadMessage(context.Background())
		if err != nil {
			slog.Error("can't read message", "err", err)
			continue
		}
		slog.Debug(op, "msg", string(msg))

		if err := n.h.ProcessMessage(msg); err != nil {
			slog.Error(op, "err", err)
			continue
		}
	}
	//
	// if err := n.nr.Close(); err != nil {
	// 	return fmt.Errorf("%s: %w", op, err)
	// }
	// return nil
}

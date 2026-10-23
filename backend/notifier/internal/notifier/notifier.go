package notifier

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"notifier/internal/receiver"
	"notifier/internal/receiver/handler"
	"notifier/internal/sender"
	"os"
	"strings"

	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"
)

type Notifier struct {
	nr   *receiver.NotifyReceiver
	sndr *sender.Sender
}

func NewNotifier(i do.Injector) (*Notifier, error) {
	const op = "NewNotifier"
	brokers := strings.Split(os.Getenv("KAFKA_PEERS"), ",")
	subCreated := os.Getenv("SUBSCRIPTION_CREATED_TOPIC")
	flightUpdated := os.Getenv("FLIGHT_UPDATED_TOPIC")
	topics := []string{subCreated, flightUpdated}
	handlers := map[string]handler.HandlerFunc{
		subCreated:    handler.SubscriptionCreatedHandler(i),
		flightUpdated: handler.FlightUpdatedHandler(i),
	}
	nr, err := receiver.NewNotifyReceiver(brokers, topics, handlers)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Notifier{
		nr:   nr,
		sndr: do.MustInvoke[*sender.Sender](i),
	}, nil
}

func (n *Notifier) Run(ctx context.Context) (result error) {
	const op = "Notifier.Run"
	defer func() {
		if err := n.nr.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("%s: close consumer: %w", op, err))
		}
	}()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := n.nr.StartConsuming(ctx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("error while consuming", "err", err)
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})

	g.Go(func() error {
		if err := n.sndr.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("error while sending", "err", err)
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Error(op, "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

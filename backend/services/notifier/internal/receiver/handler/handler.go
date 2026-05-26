package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"notifier/internal/receiver/command"
	"notifier/internal/receiver/dto"
	"notifier/internal/receiver/usecase"

	"github.com/IBM/sarama"
	"github.com/samber/do/v2"
)

type HandlerFunc func(ctx context.Context, msg *sarama.ConsumerMessage) error

func SubscriptionCreatedHandler(i do.Injector) HandlerFunc {
	uc := do.MustInvoke[*usecase.NotifierUsecase](i)
	return func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		const op = "SubscriptionCreatedHandler"
		var req dto.SubscriptionCreatedDTO
		if err := json.Unmarshal(msg.Value, &req); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		cmd, err := command.NewSubscriptionCreatedCommand(&req)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if err := uc.SaveNotification(ctx, cmd); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		slog.Debug(op + ": notification saved")
		return nil
	}
}

func FlightUpdatedHandler(i do.Injector) HandlerFunc {
	uc := do.MustInvoke[*usecase.NotifierUsecase](i)
	return func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		const op = "FlightUpdatedHandler"
		var req dto.FlightUpdatedDTO
		if err := json.Unmarshal(msg.Value, &req); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		slog.Debug(op, "req", req)

		cmd, err := command.NewFlightUpdatedCommand(&req)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if err := uc.UpdateFlight(ctx, cmd); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		slog.Debug(op + ": flight updated")
		return nil
	}
}

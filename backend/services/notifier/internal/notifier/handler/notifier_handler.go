package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"notifier/internal/infra/kafka"
	"notifier/internal/notifier/command"
	"notifier/internal/notifier/dto"
	"notifier/internal/notifier/usecase"
	"os"
	"shared/common"
)

type NotifierHandler struct {
	nr *kafka.NotifyReceiver
	uc *usecase.NotifierUsecase
}

func NewNotifierHandler(nr *kafka.NotifyReceiver) (*NotifierHandler, error) {
	const op = "NewNotifierHandler"
	if nr == nil {
		return nil, errors.New("invalid norify receiver")
	}

	appPswd := os.Getenv("APP_PASSWORD")
	appEmail, err := common.NewEmail(os.Getenv("APP_EMAIL"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	uc, err := usecase.NewNotifierUsecase(appEmail, appPswd)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &NotifierHandler{
		nr: nr,
		uc: uc,
	}, nil
}

func (h *NotifierHandler) ProcessMessage(msg []byte) error {
	const op = "NotifierHandler.ProcessMessage"
	n := dto.NotificationDTO{}
	if err := json.Unmarshal(msg, &n); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	slog.Debug(op, "NotificatioDTO", n)

	cmd, err := command.NewSendNotificationCommand(n)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := h.uc.SendNotification(cmd); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

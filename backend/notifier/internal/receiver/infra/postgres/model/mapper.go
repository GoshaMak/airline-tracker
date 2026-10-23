package model

import (
	"encoding/json"
	"notifier/internal/receiver/domain"

	"github.com/google/uuid"
)

func NotificationDomainToModel(nd domain.Notification) (NotificationModel, error) {
	carrier, err := json.Marshal(nd.TraceCarrier)
	if err != nil {
		return NotificationModel{}, err
	}
	return NotificationModel{
		Id:            nd.Id,
		Payload:       nd.Payload,
		CreatedAt:     nd.CreatedAt,
		SendAt:        nd.SendAt,
		Status:        nd.Status.String(),
		Type:          nd.Type.String(),
		EventID:       stringPointer(nd.EventID),
		Recipient:     stringPointer(nd.Recipient),
		Attempts:      nd.Attempts,
		LeaseToken:    &nd.LeaseToken,
		NextAttemptAt: nd.NextAttemptAt,
		TraceCarrier:  carrier,
	}, nil
}

func NotificationModelToDomain(nm NotificationModel) (domain.Notification, error) {
	status, err := domain.NewNotificationStatus(nm.Status)
	if err != nil {
		return domain.Notification{}, err
	}
	nt, err := domain.NewNotificationType(nm.Type)
	if err != nil {
		return domain.Notification{}, err
	}
	var carrier map[string]string
	if len(nm.TraceCarrier) > 0 {
		if err := json.Unmarshal(nm.TraceCarrier, &carrier); err != nil {
			return domain.Notification{}, err
		}
	}

	return domain.Notification{
		Id:            nm.Id,
		Payload:       nm.Payload,
		CreatedAt:     nm.CreatedAt,
		SendAt:        nm.SendAt,
		Status:        status,
		Type:          nt,
		EventID:       stringValue(nm.EventID),
		Recipient:     stringValue(nm.Recipient),
		Attempts:      nm.Attempts,
		TraceCarrier:  carrier,
		NextAttemptAt: nm.NextAttemptAt,
		LeaseToken:    uuidValue(nm.LeaseToken),
	}, nil
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func uuidValue(value *uuid.UUID) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}
	return *value
}

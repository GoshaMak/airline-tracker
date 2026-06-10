package model

import "notifier/internal/receiver/domain"

func NotificationDomainToModel(nd domain.Notification) (NotificationModel, error) {
	return NotificationModel{
		Id:        nd.Id,
		Payload:   nd.Payload,
		CreatedAt: nd.CreatedAt,
		SendAt:    nd.SendAt,
		Status:    nd.Status.String(),
		Type:      nd.Type.String(),
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

	return domain.Notification{
		Id:        nm.Id,
		Payload:   nm.Payload,
		CreatedAt: nm.CreatedAt,
		SendAt:    nm.SendAt,
		Status:    status,
		Type:      nt,
	}, nil
}

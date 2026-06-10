package model

import (
	"api/internal/publisher/domain"
)

func OutboxDomainToModel(ob domain.Outbox) (OutboxModel, error) {
	payload, err := ob.Payload.MarshalJSON()
	if err != nil {
		return OutboxModel{}, err
	}

	return OutboxModel{
		Id:        ob.Id,
		Topic:     ob.Topic,
		Payload:   payload,
		CreatedAt: ob.CreatedAt,
		SentAt:    ob.SentAt,
	}, nil
}

func OutboxModelToDomain(obm OutboxModel, pld domain.Payload) (domain.Outbox, error) {
	err := pld.UnmarshalJSON(obm.Payload)
	if err != nil {
		return domain.Outbox{}, err
	}

	return domain.Outbox{
		Id:        obm.Id,
		Topic:     obm.Topic,
		Payload:   pld,
		CreatedAt: obm.CreatedAt,
		SentAt:    obm.SentAt,
	}, nil
}

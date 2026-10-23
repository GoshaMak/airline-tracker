package model

import (
	"api/internal/publisher/domain"
	"encoding/json"

	"github.com/google/uuid"
)

func OutboxDomainToModel(ob domain.Outbox) (OutboxModel, error) {
	payload, err := ob.Payload.MarshalJSON()
	if err != nil {
		return OutboxModel{}, err
	}

	carrier, err := json.Marshal(ob.TraceCarrier)
	if err != nil {
		return OutboxModel{}, err
	}
	return OutboxModel{
		Id:            ob.Id,
		Topic:         ob.Topic,
		Payload:       payload,
		CreatedAt:     ob.CreatedAt,
		SentAt:        ob.SentAt,
		Attempts:      ob.Attempts,
		LeaseToken:    &ob.LeaseToken,
		NextAttemptAt: ob.NextAttemptAt,
		TraceCarrier:  carrier,
	}, nil
}

func OutboxModelToDomain(obm OutboxModel, pld domain.Payload) (domain.Outbox, error) {
	err := pld.UnmarshalJSON(obm.Payload)
	if err != nil {
		return domain.Outbox{}, err
	}
	var carrier map[string]string
	if len(obm.TraceCarrier) != 0 {
		if err := json.Unmarshal(obm.TraceCarrier, &carrier); err != nil {
			return domain.Outbox{}, err
		}
	}
	var leaseToken uuid.UUID
	if obm.LeaseToken != nil {
		leaseToken = *obm.LeaseToken
	}

	return domain.Outbox{
		Id:            obm.Id,
		Topic:         obm.Topic,
		Payload:       pld,
		CreatedAt:     obm.CreatedAt,
		SentAt:        obm.SentAt,
		Attempts:      obm.Attempts,
		LeaseToken:    leaseToken,
		TraceCarrier:  carrier,
		NextAttemptAt: obm.NextAttemptAt,
	}, nil
}

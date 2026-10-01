package dto

import "github.com/google/uuid"

type CreateSubscriptionRequest struct {
	FlightID            uuid.UUID `json:"flight_id" binding:"required"`
	NotifyBeforeMinutes *int64    `json:"notify_before_minutes" binding:"required,min=0"`
}

type AddSubscriptionTimersRequest struct {
	NotifyBeforeMinutes []int64 `json:"notify_before_minutes" binding:"required,min=1,unique,dive,min=0"`
}

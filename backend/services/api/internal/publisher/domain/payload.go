package domain

import "encoding/json"

type Payload interface {
	json.Marshaler
	json.Unmarshaler
}

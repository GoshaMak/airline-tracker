package domain

type Airplane struct {
	ID      uint32       `json:"id"`
	Type    AirplaneType `json:"type"`
	RowsAmt uint32       `json:"rows_amt"`
	ColsAmt uint32       `json:"cols_amt"`
}

type AirplaneType int

const (
	BoeingB707 AirplaneType = iota
	BoeingB717
	BoeingB727
	BoeingB737
	BoeingB747
	BoeingB757
	BoeingB767
	BoeingB777
	BoeingB787
)

var airplaneTypeNames = map[AirplaneType]string{
	BoeingB707: "Boeing B707",
	BoeingB717: "Boeing B717",
	BoeingB727: "Boeing B727",
	BoeingB737: "Boeing B737",
	BoeingB747: "Boeing B747",
	BoeingB757: "Boeing B757",
	BoeingB767: "Boeing B767",
	BoeingB777: "Boeing B777",
	BoeingB787: "Boeing B787",
}

func (at AirplaneType) String() string {
	return airplaneTypeNames[at]
}

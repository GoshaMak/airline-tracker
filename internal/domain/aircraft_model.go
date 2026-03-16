package domain

type AircraftModel struct {
	ID           uint
	Manufacturer string
	Model        string
	Mass         uint
	MaxAltitude  uint
	MaxSpeed     uint
}

type Model int

const (
	BoeingB707 Model = iota
	BoeingB717
	BoeingB727
	BoeingB737
	BoeingB747
	BoeingB757
	BoeingB767
	BoeingB777
	BoeingB787
)

var airplaneTypeNames = map[Model]string{
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

func (m Model) String() string {
	return airplaneTypeNames[m]
}

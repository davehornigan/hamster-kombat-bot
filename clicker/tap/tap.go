package tap

type Request struct {
	Count         int32 `json:"count"`
	Timestamp     int64 `json:"timestamp"`
	AvailableTaps int32 `json:"availableTaps"`
}

func (r *Request) IsRequest() bool {
	return true
}

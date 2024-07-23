package boost

const (
	ForBuyUri = "/boosts-for-buy"
	BuyUri    = "/buy-boost"
)

type Buy struct {
	BoostId   string `json:"boostId"`
	Timestamp int    `json:"timestamp"`
}

func (r *Buy) IsRequest() bool {
	return true
}

type Boost struct {
	Id              string `json:"id"`
	Price           int64  `json:"price"`
	EarnPerTap      int32  `json:"earnPerTap"`
	MaxTaps         int32  `json:"maxTaps"`
	CooldownSeconds int32  `json:"cooldownSeconds"`
	Level           int32  `json:"level"`
	MaxTapsDelta    int32  `json:"maxTapsDelta"`
	EarnPerTapDelta int32  `json:"earnPerTapDelta"`
	MaxLevel        int32  `json:"maxLevel,omitempty"`
}

type BoostsForBuy struct {
	Boosts []Boost `json:"boostsForBuy"`
}

package boost

import (
	"fmt"
	"github.com/dustin/go-humanize"
)

const (
	ForBuyUri = "/boosts-for-buy"
	BuyUri    = "/buy-boost"
)

type ID string

const (
	FullTaps   ID = "BoostFullAvailableTaps"
	MaxTaps    ID = "BoostMaxTaps"
	EarnPerTap ID = "BoostEarnPerTap"
)

type Buy struct {
	BoostId   ID    `json:"boostId"`
	Timestamp int64 `json:"timestamp"`
}

func (r *Buy) IsRequest() bool {
	return true
}

type Boost struct {
	Id              ID    `json:"id"`
	Price           int64 `json:"price"`
	EarnPerTap      int32 `json:"earnPerTap"`
	MaxTaps         int32 `json:"maxTaps"`
	CooldownSeconds int32 `json:"cooldownSeconds"`
	Level           int32 `json:"level"`
	MaxTapsDelta    int32 `json:"maxTapsDelta"`
	EarnPerTapDelta int32 `json:"earnPerTapDelta"`
	MaxLevel        int32 `json:"maxLevel,omitempty"`
}

func (r *Boost) Print() {
	fmt.Printf(
		"Boost: %s: %d lvl [Max %d] - Price: %s [Cooldown %d]\n",
		r.Id,
		r.Level,
		r.MaxLevel,
		humanize.Comma(r.Price),
		r.CooldownSeconds,
	)
}

type BoostsForBuy struct {
	Boosts []*Boost `json:"boostsForBuy"`
}

func (r *BoostsForBuy) IsResponse() bool {
	return true
}

package upgrades

import (
	"github.com/davehornigan/hamster-kombat-bot/clicker/user"
)

const (
	ForBuyUri = "/upgrades-for-buy"
	BuyUri    = "/buy-upgrade"
)

type UpgradeType int

const (
	MaxProfitPerCoin UpgradeType = iota
	MinCost
	MaxProfit
	None
)

type InterestingUpgradesToBuy struct {
	MaxProfit        *Upgrade
	MaxProfitPerCoin *Upgrade
	MinCost          *Upgrade
}

type Upgrade struct {
	ID                   string `json:"id"`
	Section              string `json:"section"`
	Name                 string `json:"name"`
	IsAvailable          bool   `json:"isAvailable"`
	IsExpired            bool   `json:"isExpired"`
	Price                int64  `json:"price"`
	CurrentProfitPerHour int64  `json:"currentProfitPerHour"`
	ProfitPerHour        int64  `json:"profitPerHour"`
	ProfitPerHourDelta   int64  `json:"profitPerHourDelta"`
	Level                int32  `json:"level"`
	CooldownSeconds      int32  `json:"cooldownSeconds" default:"0"`
}

func (r *Upgrade) CanBuy() bool {
	return r.IsAvailable && !r.IsExpired
}

func (r *Upgrade) GetProfitPerCoin() float64 {
	if !r.CanBuy() || r.Price == 0 || r.ProfitPerHourDelta == 0 {
		return 0
	}

	return float64(r.ProfitPerHourDelta) / float64(r.Price)
}

type Response struct {
	UpgradesForBuy []*Upgrade `json:"upgradesForBuy"`
	User           *user.User `json:"clickerUser"`
}

func (r *Response) IsResponse() bool {
	return true
}

type BuyUpgrade struct {
	UpgradeId string `json:"upgradeId"`
	Timestamp int64  `json:"timestamp"`
}

func (r *BuyUpgrade) IsRequest() bool {
	return true
}

func (r *Response) GetInterestingUpgradesAvailableToBuy() *InterestingUpgradesToBuy {
	var result *InterestingUpgradesToBuy
	for _, upgrade := range r.UpgradesForBuy {
		if !upgrade.CanBuy() {
			continue
		}
		if result == nil {
			result = &InterestingUpgradesToBuy{
				MaxProfit:        upgrade,
				MaxProfitPerCoin: upgrade,
				MinCost:          upgrade,
			}
		}
		if upgrade.ProfitPerHourDelta > result.MaxProfit.ProfitPerHourDelta {
			result.MaxProfit = upgrade
		}
		if upgrade.Price != 0 && upgrade.Price < result.MinCost.Price {
			result.MinCost = upgrade
		}
		if upgrade.GetProfitPerCoin() > result.MaxProfitPerCoin.GetProfitPerCoin() {
			result.MaxProfitPerCoin = upgrade
		}
	}

	return result
}

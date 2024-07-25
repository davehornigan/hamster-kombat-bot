package clicker

import (
	"fmt"
	"github.com/davehornigan/hamster-kombat-bot/clicker/boost"
	"github.com/davehornigan/hamster-kombat-bot/clicker/upgrades"
	"github.com/davehornigan/hamster-kombat-bot/clicker/user"
	"github.com/dustin/go-humanize"
	"time"
)

type CurrentState struct {
	AvailableTaps            int32
	MaxTaps                  int64
	TapsRecoverPerSec        int32
	Balance                  int64
	Level                    int64
	LastSyncUpdate           int64
	LastPrintState           int64
	LastCheckUpgrades        int64
	InterestingUpgradesToBuy *upgrades.InterestingUpgradesToBuy
	WantToBuyUpgrades        bool
	BuyUpgradesType          upgrades.UpgradeType
	BoostsForBuy             map[boost.ID]*boost.Boost
	WantToBuyFullTapsBoost   bool
	TapIfAvailableTaps       int32
}

func (r *CurrentState) Print() {
	fmt.Printf(
		"------\nBalance: %s\nLevel: %d\nMaxTaps: %s\n------\n",
		humanize.Comma(r.Balance),
		r.Level,
		humanize.Comma(r.MaxTaps),
	)
	if r.InterestingUpgradesToBuy != nil {
		r.InterestingUpgradesToBuy.MinCost.Print("Minimal Cost")
		r.InterestingUpgradesToBuy.MaxProfitPerCoin.Print("Max Profit Per Coin")
		r.InterestingUpgradesToBuy.MaxProfit.Print("Max Profit")
		fmt.Println("------")
	}
	if r.BoostsForBuy != nil {
		for _, boostForBuy := range r.BoostsForBuy {
			boostForBuy.Print()
		}
		fmt.Println("------")
	}
	r.LastPrintState = time.Now().Unix()
}

func (r *CurrentState) Update(clickerUser *user.User) {
	r.AvailableTaps = clickerUser.AvailableTaps
	r.MaxTaps = int64(clickerUser.MaxTaps)
	r.TapsRecoverPerSec = clickerUser.TapsRecoverPerSec
	r.Balance = int64(clickerUser.BalanceCoins)
	r.Level = clickerUser.Level
	r.LastSyncUpdate = clickerUser.LastSyncUpdate
}

func (r *CurrentState) SetBoosts(boostsForBuy []*boost.Boost) {
	res := make(map[boost.ID]*boost.Boost)
	for _, b := range boostsForBuy {
		res[b.Id] = b
	}
	r.BoostsForBuy = res
}

package clicker

import (
	"github.com/davehornigan/hamster-kombat-bot/clicker/boost"
	"github.com/davehornigan/hamster-kombat-bot/clicker/upgrades"
	"github.com/davehornigan/hamster-kombat-bot/clicker/user"
)

type CurrentState struct {
	AvailableTaps            int32
	MaxTaps                  int64
	TapsRecoverPerSec        int32
	Balance                  int64
	Level                    int64
	LastSyncUpdate           int64
	LastCheckUpgrades        int64
	InterestingUpgradesToBuy *upgrades.InterestingUpgradesToBuy
	BuyUpgradesType          upgrades.UpgradeType
	BoostsForBuy             map[boost.ID]*boost.Boost
	UseFullTapsBoost         bool
	TapIfAvailableTaps       int32
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

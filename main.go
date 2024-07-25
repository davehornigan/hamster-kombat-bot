package main

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/davehornigan/hamster-kombat-bot/clicker"
	"github.com/davehornigan/hamster-kombat-bot/clicker/boost"
	"github.com/davehornigan/hamster-kombat-bot/clicker/upgrades"
	"github.com/dustin/go-humanize"
	"log"
	"time"
)

type Config struct {
	AuthToken              string               `env:"AUTH_TOKEN"`
	UserAgent              string               `env:"USER_AGENT" envDefault:"telegram:10.14.5"`
	BuyUpgradesType        upgrades.UpgradeType `env:"BUY_UPGRADE" envDefault:"max-profit-per-coin"`
	WantToBuyUpgrades      bool                 `env:"WANT_TO_BUY_UPGRADES" envDefault:"true"`
	WantToBuyFullTapsBoost bool                 `env:"WANT_TO_BUY_FULL_TAPS" envDefault:"false"`
	TapIfAvailableTaps     int32                `env:"TAP_IF_AVAILABLE_TAPS" envDefault:"200"`
}

var clickerClient *clicker.Client
var currentState *clicker.CurrentState

func init() {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", cfg)
	clickerClient = clicker.NewClient(cfg.AuthToken, cfg.UserAgent)
	curTime := time.Now().Add(-120 * time.Second)
	currentState = &clicker.CurrentState{
		LastPrintState:         curTime.Unix(),
		LastCheckUpgrades:      curTime.Unix(),
		WantToBuyUpgrades:      cfg.WantToBuyUpgrades,
		BuyUpgradesType:        cfg.BuyUpgradesType,
		WantToBuyFullTapsBoost: cfg.WantToBuyFullTapsBoost,
		TapIfAvailableTaps:     cfg.TapIfAvailableTaps,
	}
}

func main() {
	defer handlePanic()
	clickerUser, err := clickerClient.Sync()
	if err != nil {
		log.Fatal(err)
	}
	currentState.Update(clickerUser)
	go checkAndBuyUpgrades()
	go BuyUpgrades()
	go checkBoosts()
	for {
		curTime := time.Now()
		currentState.AvailableTaps = currentState.AvailableTaps + (int32(curTime.Unix()-currentState.LastSyncUpdate) * currentState.TapsRecoverPerSec)
		currentState.LastSyncUpdate = curTime.Unix()
		if currentState.LastPrintState+30 < curTime.Unix() {
			go currentState.Print()
		}
		fmt.Printf("Available %d taps\n", currentState.AvailableTaps)
		if currentState.AvailableTaps >= currentState.TapIfAvailableTaps {
			Tap()

			if currentState.WantToBuyFullTapsBoost {
				BuyFullTapsBoost()
				Tap()
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func Tap() {
	fmt.Printf("==> Sending %d taps\n", currentState.AvailableTaps)
	clickerUser, err := clickerClient.Tap(currentState.AvailableTaps, 0)
	if err != nil {
		fmt.Printf("Error sending taps: %v\n", err)
		return
	}
	currentState.Update(clickerUser)
}

func checkBoosts() {
	boosts, err := clickerClient.GetBoostsForBuy()
	if err == nil {
		currentState.SetBoosts(boosts)
	}
}

func BuyFullTapsBoost() {
	boostForBuy := currentState.BoostsForBuy[boost.FullTaps]

	if boostForBuy.CooldownSeconds != 0 || boostForBuy.Level > boostForBuy.MaxLevel {
		return
	}

	res := make(map[boost.ID]*boost.Boost)

	boosts, err := clickerClient.BuyBoost(boostForBuy)
	if err != nil {
		for _, b := range boosts {
			res[b.Id] = b
		}
		currentState.BoostsForBuy = res
		currentState.AvailableTaps = boostForBuy.MaxTaps
		currentState.LastSyncUpdate = time.Now().Unix()
		return
	}
}

func BuyUpgrades() {
	if currentState.WantToBuyUpgrades {
		for {
			if currentState.InterestingUpgradesToBuy == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			if currentState.LastPrintState+30 < time.Now().Unix() {
				res, err := clickerClient.CheckUpgrades()
				if err != nil {
					fmt.Printf("Error checking upgrades: %v\n", err)
					return
				}
				currentState.InterestingUpgradesToBuy = res.GetInterestingUpgradesAvailableToBuy()
				currentState.LastCheckUpgrades = time.Now().Unix()
			}
			var upgradeForBuy *upgrades.Upgrade
			switch currentState.BuyUpgradesType {
			case upgrades.MaxProfitPerCoin:
				upgradeForBuy = currentState.InterestingUpgradesToBuy.MaxProfitPerCoin
			case upgrades.MinCost:
				upgradeForBuy = currentState.InterestingUpgradesToBuy.MinCost
			case upgrades.MaxProfit:
				upgradeForBuy = currentState.InterestingUpgradesToBuy.MaxProfit
			default:
				fmt.Println("No upgrades selected")
				return
			}
			if currentState.Balance < upgradeForBuy.Price {
				fmt.Printf("Insufficient funds for buy upgrade: %s\n", upgradeForBuy.Name)
				time.Sleep(3 * time.Minute)
				continue
			}
			if upgradeForBuy.CooldownSeconds > 0 {
				fmt.Printf("Wait cooldown %d sec for buy upgrade: %s\n", upgradeForBuy.CooldownSeconds, upgradeForBuy.Name)
				time.Sleep(time.Duration(upgradeForBuy.CooldownSeconds+1) * time.Second)
			}
			resp, err := clickerClient.BuyUpgrade(upgradeForBuy)
			if err != nil {
				fmt.Printf("Error buy upgrade: %s\n", err)
				continue
			}
			currentState.InterestingUpgradesToBuy = resp.GetInterestingUpgradesAvailableToBuy()
			currentState.Update(resp.User)
			fmt.Printf("^^^ Upgraded %s to %d level by price: %s\n", upgradeForBuy.Name, upgradeForBuy.Level, humanize.Comma(upgradeForBuy.Price))
		}
	}
}

func checkAndBuyUpgrades() {
	res, err := clickerClient.CheckUpgrades()
	if err != nil {
		fmt.Printf("Error checking upgrades: %v\n", err)
		return
	}
	currentState.InterestingUpgradesToBuy = res.GetInterestingUpgradesAvailableToBuy()
	currentState.LastCheckUpgrades = time.Now().Unix()
}

func handlePanic() {
	if err := recover(); err != nil {
		fmt.Println(err)
		main()
	}
}

package main

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/davehornigan/hamster-kombat-bot/clicker"
	"github.com/davehornigan/hamster-kombat-bot/clicker/upgrades"
	"github.com/davehornigan/hamster-kombat-bot/clicker/user"
	"github.com/dustin/go-humanize"
	"log"
	"time"
)

type Config struct {
	AuthToken         string               `env:"AUTH_TOKEN"`
	UserAgent         string               `env:"USER_AGENT" envDefault:"telegram:10.14.5"`
	BuyUpgradesType   upgrades.UpgradeType `env:"BUY_UPGRADE" envDefault:"max-profit-per-coin"`
	WantToBuyUpgrades bool                 `env:"WANT_TO_BUY_UPGRADES" envDefault:"true"`
}

var clickerClient *clicker.Client
var currentState *CurrentState

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
}

func init() {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", cfg)
	clickerClient = clicker.NewClient(cfg.AuthToken, cfg.UserAgent)
	curTime := time.Now().Add(-120 * time.Second)
	currentState = &CurrentState{
		LastPrintState:    curTime.Unix(),
		LastCheckUpgrades: curTime.Unix(),
		WantToBuyUpgrades: cfg.WantToBuyUpgrades,
		BuyUpgradesType:   cfg.BuyUpgradesType,
	}
}

func main() {
	defer handlePanic()
	clickerUser, err := clickerClient.Sync()
	if err != nil {
		log.Fatal(err)
	}
	updateState(clickerUser)
	go checkAndBuyUpgrades()
	go BuyUpgrades()
	for {
		curTime := time.Now()
		currentState.AvailableTaps = currentState.AvailableTaps + (int32(curTime.Unix()-currentState.LastSyncUpdate) * currentState.TapsRecoverPerSec)
		currentState.LastSyncUpdate = curTime.Unix()
		if currentState.LastPrintState+30 < curTime.Unix() {
			go printState()
		}
		fmt.Printf("Available %d taps\n", currentState.AvailableTaps)
		if currentState.AvailableTaps >= 1000 {
			go func() {
				fmt.Printf("==> Sending %d taps\n", currentState.AvailableTaps)
				clickerUser, err := clickerClient.Tap(clickerUser.AvailableTaps, 0)
				if err != nil {
					fmt.Printf("Error sending taps: %v\n", err)
					return
				}
				updateState(clickerUser)
			}()
		}
		time.Sleep(3 * time.Second)
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
			updateState(resp.User)
			fmt.Printf("^^^ Upgraded %s to %d level by price: %s\n", upgradeForBuy.Name, upgradeForBuy.Level, humanize.Comma(upgradeForBuy.Price))
		}
	}
}

func printState() {
	fmt.Printf(
		"------\nBalance: %s\nLevel: %d\nMaxTaps: %s\n------\n",
		humanize.Comma(currentState.Balance),
		currentState.Level,
		humanize.Comma(currentState.MaxTaps),
	)
	if currentState.InterestingUpgradesToBuy != nil {
		currentState.InterestingUpgradesToBuy.MinCost.Print("Minimal Cost Upgrade")
		currentState.InterestingUpgradesToBuy.MaxProfitPerCoin.Print("Max Profit Per Coin Upgrade")
		currentState.InterestingUpgradesToBuy.MaxProfit.Print("Max Profit Upgrade")
		fmt.Println("------")
	}
	currentState.LastPrintState = time.Now().Unix()
}

func updateState(clickerUser *user.User) {
	currentState.AvailableTaps = clickerUser.AvailableTaps
	currentState.MaxTaps = int64(clickerUser.MaxTaps)
	currentState.TapsRecoverPerSec = clickerUser.TapsRecoverPerSec
	currentState.Balance = int64(clickerUser.BalanceCoins)
	currentState.Level = clickerUser.Level
	currentState.LastSyncUpdate = clickerUser.LastSyncUpdate
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

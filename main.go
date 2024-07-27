package main

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/davehornigan/hamster-kombat-bot/clicker"
	"github.com/davehornigan/hamster-kombat-bot/clicker/boost"
	"github.com/davehornigan/hamster-kombat-bot/clicker/upgrades"
	"github.com/davehornigan/hamster-kombat-bot/view"
	"github.com/dustin/go-humanize"
	"log"
	"os"
	"time"
)

type Config struct {
	AuthToken string `env:"AUTH_TOKEN"`
	UserAgent string `env:"USER_AGENT" envDefault:"telegram:10.14.5"`
}

var clickerClient *clicker.Client
var currentState *clicker.CurrentState
var printer *view.Printer

func init() {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}
	clickerClient = clicker.NewClient(cfg.AuthToken, cfg.UserAgent)
	curTime := time.Now().Add(-120 * time.Second)
	currentState = &clicker.CurrentState{
		LastCheckUpgrades:  curTime.Unix(),
		BuyUpgradesType:    upgrades.None,
		UseFullTapsBoost:   false,
		TapIfAvailableTaps: 500,
	}
	printer = view.NewPrinter()
}

func main() {
	app := printer.InitApp(currentState)

	go func() {
		if err := app.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			printer.UpdateDraw()
		}
	}()

	Sync()
	checkUpgrades()

	go func() {
		go BuyUpgrades()
		go checkBoosts()
		for {
			curTime := time.Now()
			currentState.AvailableTaps = currentState.AvailableTaps + (int32(curTime.Unix()-currentState.LastSyncUpdate) * currentState.TapsRecoverPerSec)
			currentState.LastSyncUpdate = curTime.Unix()
			if currentState.AvailableTaps >= currentState.TapIfAvailableTaps {
				Tap()

				if currentState.UseFullTapsBoost {
					BuyFullTapsBoost()
				}
			}
		}
	}()

	select {}
}

func Sync() {
	clickerUser, err := clickerClient.Sync()
	if err != nil {
		log.Fatal(err)
	}
	currentState.Update(clickerUser)
}

func Tap() {
	printer.LogLine(fmt.Sprintf("Sending %d taps", currentState.AvailableTaps))
	clickerUser, err := clickerClient.Tap(currentState.AvailableTaps, 0)
	if err != nil {
		printer.LogLine(fmt.Sprintf("Error sending taps: %v\n", err))
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
	if currentState.BoostsForBuy == nil {
		return
	}
	boostForBuy := currentState.BoostsForBuy[boost.FullTaps]

	if boostForBuy.CooldownSeconds != 0 || boostForBuy.Level > boostForBuy.MaxLevel {
		go checkBoosts()
		return
	}

	res := make(map[boost.ID]*boost.Boost)

	boosts, err := clickerClient.BuyBoost(boostForBuy)
	if err == nil {
		for _, b := range boosts {
			res[b.Id] = b
		}
		currentState.BoostsForBuy = res
		currentState.AvailableTaps = boostForBuy.MaxTaps
		currentState.LastSyncUpdate = time.Now().Unix()

		Sync()
	}
}

func BuyUpgrades() {
	for {
		if currentState.BuyUpgradesType == upgrades.None {
			time.Sleep(30 * time.Second)
			continue
		}
		if currentState.InterestingUpgradesToBuy == nil {
			time.Sleep(1 * time.Second)
			continue
		}
		if currentState.LastCheckUpgrades+30 < time.Now().Unix() {
			res, err := clickerClient.CheckUpgrades()
			if err != nil {
				printer.LogLine(fmt.Sprintf("Error checking upgrades: %v\n", err))
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
			return
		}
		if currentState.Balance < upgradeForBuy.Price {
			printer.LogLine(fmt.Sprintf("Insufficient funds for buy upgrade \"%s\". Wait 3 min.", upgradeForBuy.Name))
			time.Sleep(3 * time.Minute)
			continue
		}
		if upgradeForBuy.CooldownSeconds > 0 {
			printer.LogLine(fmt.Sprintf("Wait cooldown %d sec for buy upgrade \"%s\"", upgradeForBuy.CooldownSeconds, upgradeForBuy.Name))
			time.Sleep(time.Duration(upgradeForBuy.CooldownSeconds+1) * time.Second)
		}
		resp, err := clickerClient.BuyUpgrade(upgradeForBuy)
		if err != nil {
			printer.LogLine(fmt.Sprintf("Error buy upgrade: %s\n", err))
			continue
		}
		currentState.InterestingUpgradesToBuy = resp.GetInterestingUpgradesAvailableToBuy()
		currentState.Update(resp.User)
		printer.LogLine(fmt.Sprintf("Upgraded \"%s\" to %d level by price: %s", upgradeForBuy.Name, upgradeForBuy.Level, humanize.Comma(upgradeForBuy.Price)))
	}
}

func checkUpgrades() {
	res, err := clickerClient.CheckUpgrades()
	if err != nil {
		printer.LogLine(fmt.Sprintf("Error checking upgrades: %v", err))
		return
	}
	currentState.InterestingUpgradesToBuy = res.GetInterestingUpgradesAvailableToBuy()
	currentState.LastCheckUpgrades = time.Now().Unix()
}

package view

import (
	"fmt"
	"github.com/davehornigan/hamster-kombat-bot/clicker"
	"github.com/davehornigan/hamster-kombat-bot/clicker/boost"
	"github.com/davehornigan/hamster-kombat-bot/clicker/upgrades"
	"github.com/dustin/go-humanize"
	"github.com/rivo/tview"
	"log"
	"strconv"
	"strings"
	"unicode"
)

type Printer struct {
	app        *tview.Application
	UpdateDraw func()
	LogLine    func(string)
}

func NewPrinter() *Printer {
	return &Printer{
		app: tview.NewApplication(),
	}
}

func (r *Printer) InitApp(currentState *clicker.CurrentState) *tview.Application {
	upgradesTextView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	upgradesTextView.SetTitle("Upgrades").SetBorder(true)
	boostsTextView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	boostsTextView.SetTitle("Boosts").SetBorder(true)

	tapsView := tview.NewTextView().SetLabel("Taps available: ").SetDynamicColors(true)
	maxTapsView := tview.NewTextView().SetLabel("Max Taps: ").SetDynamicColors(true)
	tapsRecoverView := tview.NewTextView().SetLabel("Taps Recover: ").SetDynamicColors(true)
	levelView := tview.NewTextView().SetLabel("Level: ").SetDynamicColors(true)
	balanceView := tview.NewTextView().SetLabel("Balance: ").SetDynamicColors(true)

	logTextView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	logTextView.SetTitle("Log").SetBorder(true)

	r.LogLine = func(text string) {
		r.app.QueueUpdateDraw(func() {
			lines := strings.Split(logTextView.GetText(false), "\n")
			if len(lines) >= 20 {
				lines = lines[1:]
			}
			lines = append(lines, text)
			logTextView.SetText(strings.Join(lines, "\n"))
		})
	}

	r.UpdateDraw = func() {
		r.app.QueueUpdateDraw(func() {
			r.DrawUpgrades(upgradesTextView, currentState.InterestingUpgradesToBuy)
			r.DrawBoosts(boostsTextView, currentState.BoostsForBuy)
			balanceView.SetText(fmt.Sprintf("%s", humanize.Comma(currentState.Balance)))
			levelView.SetText(fmt.Sprintf("%d", currentState.Level))
			tapsView.SetText(fmt.Sprintf("%s", humanize.Comma(int64(currentState.AvailableTaps))))
			maxTapsView.SetText(fmt.Sprintf("%s", humanize.Comma(currentState.MaxTaps)))
			tapsRecoverView.SetText(fmt.Sprintf("%d/sec.", currentState.TapsRecoverPerSec))
		})
	}

	form := tview.NewForm()
	form.SetBorder(true)
	form.AddDropDown("Buy Upgrades by strategy", []string{"Max Profit Per Coin", "Minimal Cost", "Max Profit", "None"}, 3, func(option string, optionIndex int) {
		currentState.BuyUpgradesType = upgrades.UpgradeType(optionIndex)
	})
	form.AddCheckbox("Use Full Energy booster", false, func(checked bool) {
		currentState.UseFullTapsBoost = checked
	})
	form.AddInputField(
		"Send taps when accumulated:",
		strconv.Itoa(int(currentState.TapIfAvailableTaps)),
		10,
		func(textToCheck string, lastChar rune) bool {
			return unicode.IsDigit(lastChar)
		},
		func(text string) {
			taps, err := strconv.ParseInt(text, 10, 32)
			if err != nil {
				log.Fatal(err)
			}

			currentState.TapIfAvailableTaps = int32(taps)
		})

	tapsFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tapsView, 0, 1, false).
		AddItem(tapsRecoverView, 0, 1, false).
		AddItem(maxTapsView, 0, 1, false)
	userInfoFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(balanceView, 0, 1, false).
		AddItem(levelView, 0, 1, false)

	stats := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tapsFlex, 0, 1, false).
		AddItem(userInfoFlex, 0, 1, false).
		AddItem(upgradesTextView, 0, 1, false).
		AddItem(boostsTextView, 0, 1, false)
	stats.SetBorder(true)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(form, 0, 1, true).
		AddItem(stats, 0, 1, false).
		AddItem(logTextView, 0, 1, false)

	r.app.SetRoot(flex, true)

	return r.app
}

func (r *Printer) DrawUpgrades(view *tview.TextView, upgrades *upgrades.InterestingUpgradesToBuy) {
	lines := make([]string, 0)
	if upgrades != nil {
		format := "%s: %s > %s [%d lvl] - %s [+%s per hour]\n"
		lines = append(lines, fmt.Sprintf(
			format,
			"Max Profit Per Coin",
			upgrades.MaxProfitPerCoin.Section,
			upgrades.MaxProfitPerCoin.Name,
			upgrades.MaxProfitPerCoin.Level,
			humanize.Comma(upgrades.MaxProfitPerCoin.Price),
			humanize.Comma(upgrades.MaxProfitPerCoin.ProfitPerHourDelta),
		))
		lines = append(lines, fmt.Sprintf(
			format,
			"Max Profit",
			upgrades.MaxProfit.Section,
			upgrades.MaxProfit.Name,
			upgrades.MaxProfit.Level,
			humanize.Comma(upgrades.MaxProfit.Price),
			humanize.Comma(upgrades.MaxProfit.ProfitPerHourDelta),
		))
		lines = append(lines, fmt.Sprintf(
			format,
			"Min Cost",
			upgrades.MinCost.Section,
			upgrades.MinCost.Name,
			upgrades.MinCost.Level,
			humanize.Comma(upgrades.MinCost.Price),
			humanize.Comma(upgrades.MinCost.ProfitPerHourDelta),
		))
	}
	if len(lines) == 0 {
		lines = append(lines, "Wait upgrades check...")
	}
	view.SetText(strings.Join(lines[:], ""))
}

func (r *Printer) DrawBoosts(view *tview.TextView, boosts map[boost.ID]*boost.Boost) {
	lines := make([]string, 0)
	for _, boostForBuy := range boosts {
		lines = append(lines, fmt.Sprintf(
			"%s: %d lvl [Max %d] - Price: %s [Cooldown %d]\n",
			boostForBuy.Id,
			boostForBuy.Level,
			boostForBuy.MaxLevel,
			humanize.Comma(boostForBuy.Price),
			boostForBuy.CooldownSeconds,
		))
		if len(lines) == 0 {
			lines = append(lines, "Wait boosts check...")
		}
		view.SetText(strings.Join(lines[:], ""))
	}
}

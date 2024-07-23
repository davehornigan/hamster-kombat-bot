package user

type User struct {
	ID                string  `json:"id"`
	TotalCoins        float64 `json:"totalCoins"`
	BalanceCoins      float64 `json:"balanceCoins"`
	Level             int64   `json:"level"`
	AvailableTaps     int32   `json:"availableTaps"`
	MaxTaps           int32   `json:"maxTaps"`
	TapsRecoverPerSec int32   `json:"tapsRecoverPerSec"`
	LastSyncUpdate    int64   `json:"lastSyncUpdate"`
}

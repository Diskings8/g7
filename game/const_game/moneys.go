package const_game

const (
	CurrencyType_Gold    int8 = iota + 1 // 金币
	CurrencyType_Diamond                 // 钻石
	CurrencyType_Ticket                  // 点券/抽奖券
	CurrencyType_Energy                  // 体力
)

const (
	BagType_Default uint8 = iota
	BagType_Currency
)

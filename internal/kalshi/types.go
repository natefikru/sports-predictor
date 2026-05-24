package kalshi

type ScopeGroup int

const (
	ScopeMoneyline ScopeGroup = iota
	ScopeGameSpread
	ScopeGameTotal
	Scope1stHalfTotal
	Scope1stHalfSpread
	Scope2ndHalfSpread
	ScopeQuarterMarkets
	ScopeSeriesSpread
	ScopeSeriesWinner
	ScopeSeriesGames
	ScopeSeriesExact
	ScopeSeriesGame7
	ScopePlayerPoints
	ScopePlayer3pt
	ScopePlayerRebounds
	ScopePlayerAssists
	ScopePlayerSteals
	ScopePlayerCombined  // PRA, PR, PA, RA combos
	ScopePlayerDoubleDouble
	ScopeUnknown
)

var ScopeLabel = map[ScopeGroup]string{
	ScopeMoneyline:          "MONEYLINE",
	ScopeGameSpread:         "GAME SPREAD",
	ScopeGameTotal:          "GAME TOTAL",
	Scope1stHalfTotal:       "1ST HALF TOTAL",
	Scope1stHalfSpread:      "1ST HALF SPREAD",
	Scope2ndHalfSpread:      "2ND HALF SPREAD",
	ScopeQuarterMarkets:     "QUARTER MARKETS",
	ScopeSeriesSpread:       "SERIES GAME SPREAD",
	ScopeSeriesWinner:       "SERIES WINNER",
	ScopeSeriesGames:        "SERIES TOTAL GAMES",
	ScopeSeriesExact:        "SERIES EXACT SCORE",
	ScopeSeriesGame7:        "GOES TO GAME 7?",
	ScopePlayerPoints:       "PLAYER POINTS",
	ScopePlayer3pt:          "PLAYER 3-POINTERS",
	ScopePlayerRebounds:     "PLAYER REBOUNDS",
	ScopePlayerAssists:      "PLAYER ASSISTS",
	ScopePlayerSteals:       "PLAYER STEALS",
	ScopePlayerCombined:     "PLAYER COMBINED STATS",
	ScopePlayerDoubleDouble: "DOUBLE-DOUBLE PROPS",
}

// ScopeDisplayOrder controls the render order of sections.
// Game markets first, then series-level, then player props.
var ScopeDisplayOrder = []ScopeGroup{
	ScopeMoneyline, ScopeGameSpread, ScopeGameTotal,
	Scope1stHalfTotal, Scope1stHalfSpread, Scope2ndHalfSpread, ScopeQuarterMarkets,
	ScopeSeriesWinner, ScopeSeriesSpread, ScopeSeriesGames,
	ScopeSeriesExact, ScopeSeriesGame7,
	ScopePlayerPoints, ScopePlayer3pt, ScopePlayerRebounds,
	ScopePlayerAssists, ScopePlayerSteals, ScopePlayerCombined, ScopePlayerDoubleDouble,
}

// isGameLevel returns true for scopes tied to a specific game date.
func isGameLevel(g ScopeGroup) bool {
	switch g {
	case ScopeMoneyline, ScopeGameSpread, ScopeGameTotal,
		Scope1stHalfTotal, Scope1stHalfSpread, Scope2ndHalfSpread, ScopeQuarterMarkets:
		return true
	}
	return false
}

// isPlayerProp returns true for player-prop scopes.
func isPlayerProp(g ScopeGroup) bool {
	switch g {
	case ScopePlayerPoints, ScopePlayer3pt, ScopePlayerRebounds,
		ScopePlayerAssists, ScopePlayerSteals, ScopePlayerCombined, ScopePlayerDoubleDouble:
		return true
	}
	return false
}

type Series struct {
	SeriesTicker string   `json:"ticker"`
	Title        string   `json:"title"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
}

type SeriesScope struct {
	Ticker string
	Title  string
	Group  ScopeGroup
}

type Event struct {
	EventTicker  string `json:"event_ticker"`
	SeriesTicker string `json:"series_ticker"`
	Title        string `json:"title"`
	SubTitle     string `json:"sub_title"`
	TeamA        string `json:"-"`
	TeamB        string `json:"-"`
}

type Market struct {
	Ticker         string `json:"ticker"`
	EventTicker    string `json:"event_ticker"`
	Title          string `json:"title"`
	YesSubTitle    string `json:"yes_sub_title"`
	YesBid         string `json:"yes_bid_dollars"`
	YesAsk         string `json:"yes_ask_dollars"`
	NoBid          string `json:"no_bid_dollars"`
	NoAsk          string `json:"no_ask_dollars"`
	LastPrice      string `json:"last_price_dollars"`
	VolumeFP       string `json:"volume_fp"`
	Volume24hFP    string `json:"volume_24h_fp"`
	OpenInterestFP string `json:"open_interest_fp"`
	CloseTime      string `json:"close_time"`
	Status         string `json:"status"`
}

type MarketSection struct {
	Scope    ScopeGroup
	Label    string
	Markets  []Market
	FetchErr bool
}

type GameView struct {
	GameEvent Event
	Sections  []MarketSection
}

type EventsResponse struct {
	Events []Event `json:"events"`
	Cursor string  `json:"cursor"`
}

type MarketsResponse struct {
	Markets []Market `json:"markets"`
	Cursor  string   `json:"cursor"`
}

type SeriesResponse struct {
	Series []Series `json:"series"`
	Cursor string   `json:"cursor"`
}

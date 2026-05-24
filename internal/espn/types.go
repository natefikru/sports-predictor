package espn

// TeamInfo holds an ESPN team's identity.
type TeamInfo struct {
	ID           string
	Abbreviation string
	DisplayName  string
	ShortName    string
}

// GameResult is a single completed game in the series.
type GameResult struct {
	GameID    string
	Date      string
	HomeTeam  TeamInfo
	AwayTeam  TeamInfo
	HomeScore int
	AwayScore int
	Status    string // "STATUS_FINAL" | "STATUS_SCHEDULED"
}

// PlayerGameLine is one player's box score for a single game.
type PlayerGameLine struct {
	GameID  string
	Points  float64
	Rebounds float64
	Assists  float64
	Steals   float64
	Blocks   float64
	Minutes  float64
	ThreePM  float64
	ThreePA  float64
	FGM      float64
	FGA      float64
	FTM      float64
	FTA      float64
}

// PlayerSeries aggregates a player's stats across all fetched series games.
type PlayerSeries struct {
	PlayerID    string
	DisplayName string
	TeamID      string
	Lines       []PlayerGameLine
	Avg         PlayerGameLine // computed averages
}

// SeriesReport is the complete data package for one playoff series.
type SeriesReport struct {
	TeamA        TeamInfo
	TeamB        TeamInfo
	SeriesRecord string // e.g. "OKC leads 2-1"
	Games        []GameResult
	Players      []*PlayerSeries // all players with meaningful minutes
	AvgTotal     float64         // avg combined points per game
}

// rosterEntry is used when parsing the ESPN core roster response.
type rosterEntry struct {
	PlayerID    int    `json:"playerId"`
	DisplayName string `json:"displayName"`
	DidNotPlay  bool   `json:"didNotPlay"`
	Statistics  struct {
		Ref string `json:"$ref"`
	} `json:"statistics"`
}

type rosterResponse struct {
	Entries []rosterEntry `json:"entries"`
}

// statCategory groups named stat values.
type statCategory struct {
	Name  string    `json:"name"`
	Stats []statVal `json:"stats"`
}

type statVal struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type playerStatResponse struct {
	Splits struct {
		Categories []statCategory `json:"categories"`
	} `json:"splits"`
}

// scheduleEvent is a lite event from the team schedule response.
type scheduleEvent struct {
	ID           string `json:"id"`
	Date         string `json:"date"`
	Competitions []struct {
		Status struct {
			Type struct {
				Name string `json:"name"`
			} `json:"type"`
		} `json:"status"`
		Competitors []struct {
			ID       string `json:"id"`
			HomeAway string `json:"homeAway"`
			Score    struct {
				DisplayValue string `json:"displayValue"`
			} `json:"score"`
			Winner bool `json:"winner"`
			Team   struct {
				ID               string `json:"id"`
				Abbreviation     string `json:"abbreviation"`
				DisplayName      string `json:"displayName"`
				ShortDisplayName string `json:"shortDisplayName"`
			} `json:"team"`
		} `json:"competitors"`
	} `json:"competitions"`
}

type scheduleResponse struct {
	Events []scheduleEvent `json:"events"`
}

// espnTeam is used when reading the teams list.
type espnTeam struct {
	ID               string `json:"id"`
	Abbreviation     string `json:"abbreviation"`
	DisplayName      string `json:"displayName"`
	ShortDisplayName string `json:"shortDisplayName"`
}

type teamsResponse struct {
	Sports []struct {
		Leagues []struct {
			Teams []struct {
				Team espnTeam `json:"team"`
			} `json:"teams"`
		} `json:"leagues"`
	} `json:"sports"`
}

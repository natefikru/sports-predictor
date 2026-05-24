package kalshi

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// DiscoverAndClassifyNBASeries fetches all Basketball-tagged series from Kalshi
// and classifies each into a ScopeGroup by title keyword matching.
// Excluded categories (futures, awards) are omitted from the result.
func DiscoverAndClassifyNBASeries(c *Client) ([]SeriesScope, error) {
	body, err := c.get("/series", map[string]string{
		"tags":  "Basketball",
		"limit": "200",
	})
	if err != nil {
		return nil, fmt.Errorf("fetching series: %w", err)
	}

	var resp SeriesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing series: %w", err)
	}

	var scopes []SeriesScope
	for _, s := range resp.Series {
		// Skip non-NBA leagues: WNBA (KXWNBA*), NCAA (KXNCAAM*, KXNCAAW*)
		// and other basketball leagues that match the Basketball tag.
		if isNonNBATicker(s.SeriesTicker) {
			continue
		}
		g := classifySeriesTitle(s.Title)
		if g == ScopeUnknown {
			continue
		}
		scopes = append(scopes, SeriesScope{
			Ticker: s.SeriesTicker,
			Title:  s.Title,
			Group:  g,
		})
	}
	return scopes, nil
}

// classifySeriesTitle maps a series title to a ScopeGroup.
// Returns ScopeUnknown for excluded or unrecognized titles.
func classifySeriesTitle(title string) ScopeGroup {
	t := strings.ToLower(title)

	// Excluded categories — return ScopeUnknown to skip entirely
	for _, excl := range []string{"championship", " mvp", "award", "rookie", "coach", "futures", "slam dunk", "3-point contest", "all-star", "celebrity", "draft", "win total"} {
		if strings.Contains(t, excl) {
			return ScopeUnknown
		}
	}

	switch {
	// Game-level markets
	case strings.Contains(t, "professional basketball game"):
		return ScopeMoneyline
	case t == "pro basketball spread":
		return ScopeGameSpread
	case t == "pro basketball total points":
		return ScopeGameTotal
	case strings.Contains(t, "team total"):
		return ScopeGameTotal // group team totals with game total
	case strings.Contains(t, "1st half total") || strings.Contains(t, "1h total"):
		return Scope1stHalfTotal
	case strings.Contains(t, "1st half spread") || strings.Contains(t, "1h spread") || strings.Contains(t, "1st half winner"):
		return Scope1stHalfSpread
	case strings.Contains(t, "2nd half spread") || strings.Contains(t, "2h spread") ||
		strings.Contains(t, "2nd half total") || strings.Contains(t, "2nd half winner"):
		return Scope2ndHalfSpread
	case strings.Contains(t, "quarter"):
		return ScopeQuarterMarkets

	// Series-level markets
	case strings.Contains(t, "series game spread"):
		return ScopeSeriesSpread
	case strings.Contains(t, "series winner") || strings.Contains(t, "professional basketball series") && !strings.Contains(t, "women"):
		return ScopeSeriesWinner
	case strings.Contains(t, "series total") || strings.Contains(t, "total games"):
		return ScopeSeriesGames
	case strings.Contains(t, "series exact") || strings.Contains(t, "championship series score"):
		return ScopeSeriesExact
	case strings.Contains(t, "game 7"):
		return ScopeSeriesGame7

	// Player props — order matters: check combined stats before individual
	case strings.Contains(t, "points + rebounds + assists") || strings.Contains(t, "head-to-head combined"):
		return ScopePlayerCombined
	case strings.Contains(t, "points + rebounds"):
		return ScopePlayerCombined
	case strings.Contains(t, "points + assists"):
		return ScopePlayerCombined
	case strings.Contains(t, "rebounds + assists"):
		return ScopePlayerCombined
	case strings.Contains(t, "head-to-head points") || (strings.Contains(t, "player points") && !strings.Contains(t, "leader")):
		return ScopePlayerPoints
	case strings.Contains(t, "player threes") || strings.Contains(t, "three pointers") || strings.Contains(t, "head-to-head three"):
		return ScopePlayer3pt
	case strings.Contains(t, "player rebounds"):
		return ScopePlayerRebounds
	case strings.Contains(t, "player assists"):
		return ScopePlayerAssists
	case strings.Contains(t, "player steals"):
		return ScopePlayerSteals
	case strings.Contains(t, "player blocks"):
		return ScopePlayerCombined // group blocks with combined
	case strings.Contains(t, "double double") || strings.Contains(t, "double-double"):
		return ScopePlayerDoubleDouble
	case strings.Contains(t, "triple double"):
		return ScopePlayerDoubleDouble // group triple-double with double-double

	default:
		return ScopeUnknown
	}
}

// FetchGameEvents fetches open game-winner events for the moneyline series ticker.
// These are what the user picks from.
func FetchGameEvents(c *Client, seriesTicker string) ([]Event, error) {
	body, err := c.get("/events", map[string]string{
		"status":        "open",
		"series_ticker": seriesTicker,
		"limit":         "50",
	})
	if err != nil {
		return nil, fmt.Errorf("fetching game events: %w", err)
	}

	var resp EventsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing game events: %w", err)
	}

	for i := range resp.Events {
		a, b, _ := ExtractTeams(resp.Events[i].EventTicker, resp.Events[i].SeriesTicker, resp.Events[i].SubTitle)
		resp.Events[i].TeamA = a
		resp.Events[i].TeamB = b
	}
	return resp.Events, nil
}

// FetchMarketsForEvent fetches all active markets for a specific event ticker.
func FetchMarketsForEvent(c *Client, eventTicker string) ([]Market, error) {
	body, err := c.get("/markets", map[string]string{
		"status":       "open",
		"event_ticker": eventTicker,
		"limit":        "100",
	})
	if err != nil {
		return nil, fmt.Errorf("fetching markets for %s: %w", eventTicker, err)
	}

	var resp MarketsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing markets for %s: %w", eventTicker, err)
	}

	// Filter to active only
	var active []Market
	for _, m := range resp.Markets {
		if m.Status == "active" || m.Status == "" {
			active = append(active, m)
		}
	}
	return active, nil
}

// isNonNBATicker returns true for series tickers that belong to non-NBA leagues
// (WNBA, NCAA, etc.) that share the Basketball tag but aren't Pro Basketball (M).
func isNonNBATicker(ticker string) bool {
	nonNBAPrefixes := []string{
		"KXWNBA",    // WNBA
		"KXNCAAMB",  // NCAA Men's Basketball
		"KXNCAAW",   // NCAA Women's Basketball
		"KXEUROBA",  // EuroLeague
		"KXNBLAU",   // Australian NBL
		"KXCBA",     // Chinese Basketball Association
	}
	upper := strings.ToUpper(ticker)
	for _, prefix := range nonNBAPrefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

// FetchAllScopes fetches open events for all scope groups sequentially with a small
// delay between requests to respect Kalshi's rate limits.
func FetchAllScopes(c *Client, scopes []SeriesScope) map[ScopeGroup][]Event {
	result := make(map[ScopeGroup][]Event)

	for _, s := range scopes {
		body, err := c.get("/events", map[string]string{
			"status":        "open",
			"series_ticker": s.Ticker,
			"limit":         "50",
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [warn] %s: %v\n", s.Ticker, err)
			time.Sleep(500 * time.Millisecond) // back off on error
			continue
		}
		var resp EventsResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			continue
		}
		for i := range resp.Events {
			a, b, _ := ExtractTeams(resp.Events[i].EventTicker, resp.Events[i].SeriesTicker, resp.Events[i].SubTitle)
			resp.Events[i].TeamA = a
			resp.Events[i].TeamB = b
		}
		result[s.Group] = append(result[s.Group], resp.Events...)
		time.Sleep(100 * time.Millisecond)
	}

	return result
}

// MatchEventsToGame filters a pool of all-scope events to those pertaining
// to the same game as the selected event.
// Game-level scopes require teams + date match; series-level require teams only;
// player props require at least one team match.
func MatchEventsToGame(selected Event, pool map[ScopeGroup][]Event) map[ScopeGroup][]Event {
	dateFragment, hasDate := ExtractDateFragment(selected.EventTicker, selected.SeriesTicker)
	matched := make(map[ScopeGroup][]Event)

	for scope, events := range pool {
		for _, e := range events {
			a, b := e.TeamA, e.TeamB
			sameTeams := teamsMatch(selected.TeamA, selected.TeamB, a, b)

			if isPlayerProp(scope) {
				// Player props only need one of the two teams
				if containsTeam(selected.TeamA, selected.TeamB, a) || containsTeam(selected.TeamA, selected.TeamB, b) {
					matched[scope] = append(matched[scope], e)
				}
				continue
			}

			if isGameLevel(scope) {
				// Game-level: teams + date must match
				candidateDate, _ := ExtractDateFragment(e.EventTicker, e.SeriesTicker)
				if sameTeams && hasDate && candidateDate == dateFragment {
					matched[scope] = append(matched[scope], e)
				}
				continue
			}

			// Series-level: teams only
			if sameTeams {
				matched[scope] = append(matched[scope], e)
			}
		}
	}
	return matched
}

// FetchMarketsForAllSections fetches markets for all matched events sequentially
// and returns them as ordered MarketSections.
func FetchMarketsForAllSections(c *Client, matched map[ScopeGroup][]Event, moneylineMarkets []Market) []MarketSection {
	type scopeResult struct {
		markets  []Market
		fetchErr bool
	}

	results := make(map[ScopeGroup]*scopeResult)
	results[ScopeMoneyline] = &scopeResult{markets: moneylineMarkets}

	for scope, events := range matched {
		if scope == ScopeMoneyline {
			continue
		}
		r := &scopeResult{}
		results[scope] = r
		for _, e := range events {
			markets, err := FetchMarketsForEvent(c, e.EventTicker)
			if err != nil {
				r.fetchErr = true
				continue
			}
			r.markets = append(r.markets, markets...)
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Check moneyline markets for OT outcomes and extract to a synthetic section
	var otMarkets []Market
	var filteredMoneyline []Market
	for _, m := range moneylineMarkets {
		label := strings.ToLower(m.YesSubTitle + " " + m.Title)
		if strings.Contains(label, "overtime") || strings.Contains(label, " ot ") || strings.Contains(label, " ot?") {
			otMarkets = append(otMarkets, m)
		} else {
			filteredMoneyline = append(filteredMoneyline, m)
		}
	}
	if results[ScopeMoneyline] != nil {
		results[ScopeMoneyline].markets = filteredMoneyline
	}

	// Build ordered sections
	var sections []MarketSection
	for _, scope := range ScopeDisplayOrder {
		r, ok := results[scope]
		if !ok {
			r = &scopeResult{}
		}
		sections = append(sections, MarketSection{
			Scope:    scope,
			Label:    ScopeLabel[scope],
			Markets:  r.markets,
			FetchErr: r.fetchErr,
		})
	}

	// Insert OT section after quarter markets if we found OT markets
	if len(otMarkets) > 0 {
		otSection := MarketSection{
			Scope:   ScopeUnknown,
			Label:   "OVERTIME",
			Markets: otMarkets,
		}
		for i, s := range sections {
			if s.Scope == ScopeQuarterMarkets {
				newSections := make([]MarketSection, 0, len(sections)+1)
				newSections = append(newSections, sections[:i+1]...)
				newSections = append(newSections, otSection)
				newSections = append(newSections, sections[i+1:]...)
				sections = newSections
				break
			}
		}
	}

	return sections
}

// BuildGameView is the main orchestrator: discovers scopes, fetches all events
// concurrently, matches to the selected game, fetches markets, and assembles a GameView.
func BuildGameView(c *Client, selected Event, allScopes []SeriesScope) (GameView, error) {
	// Fetch moneyline markets for the selected event first
	moneylineMarkets, err := FetchMarketsForEvent(c, selected.EventTicker)
	if err != nil {
		return GameView{}, fmt.Errorf("fetching game markets: %w", err)
	}

	// Fetch events for all other scope groups concurrently
	otherScopes := make([]SeriesScope, 0, len(allScopes))
	for _, s := range allScopes {
		if s.Group != ScopeMoneyline {
			otherScopes = append(otherScopes, s)
		}
	}
	allEvents := FetchAllScopes(c, otherScopes)

	// Match events to this specific game
	matched := MatchEventsToGame(selected, allEvents)

	// Fetch markets for all matched events and assemble sections
	sections := FetchMarketsForAllSections(c, matched, moneylineMarkets)

	return GameView{
		GameEvent: selected,
		Sections:  sections,
	}, nil
}

// months is the set of valid 3-letter month abbreviations used in Kalshi tickers.
var months = map[string]bool{
	"JAN": true, "FEB": true, "MAR": true, "APR": true,
	"MAY": true, "JUN": true, "JUL": true, "AUG": true,
	"SEP": true, "OCT": true, "NOV": true, "DEC": true,
}

// ExtractTeams parses the two team abbreviations from a Kalshi event ticker.
//
// Kalshi ticker formats after stripping the series prefix:
//   - Game-level:   "26MAY26SASOKC"  (2-digit year + 3-letter month + 2-digit day + teamA + teamB)
//   - Series-level: "26SASOKCWCF"    (2-digit year + teamA + teamB + optional suffix)
//
// Falls back to sub_title parsing ("SAS at OKC") if ticker parsing fails.
func ExtractTeams(eventTicker, seriesTicker, subTitle string) (teamA, teamB string, ok bool) {
	prefix := seriesTicker + "-"
	remainder := strings.TrimPrefix(eventTicker, prefix)
	if remainder == eventTicker {
		return extractTeamsFromSubTitle(subTitle)
	}

	i := 0

	// Skip leading 2-digit year if present
	if len(remainder) >= 2 && isDigit(remainder[0]) && isDigit(remainder[1]) {
		i = 2
	}

	// Skip 3-letter month + 2-digit day if this is a game-level ticker
	if i+5 <= len(remainder) && months[remainder[i:i+3]] && isDigit(remainder[i+3]) && isDigit(remainder[i+4]) {
		i += 5
	}

	// Need at least 6 chars for two 3-letter team codes
	if i+6 > len(remainder) {
		return extractTeamsFromSubTitle(subTitle)
	}

	teamA = remainder[i : i+3]
	teamB = remainder[i+3 : i+6]

	if !isUpperAlpha(teamA) || !isUpperAlpha(teamB) {
		return extractTeamsFromSubTitle(subTitle)
	}
	return teamA, teamB, true
}

// ExtractDateFragment returns the date component (e.g. "26MAY26") from a game event ticker.
func ExtractDateFragment(eventTicker, seriesTicker string) (string, bool) {
	prefix := seriesTicker + "-"
	remainder := strings.TrimPrefix(eventTicker, prefix)
	if remainder == eventTicker {
		return "", false
	}

	i := 0
	// 2-digit year
	if len(remainder) >= 2 && isDigit(remainder[0]) && isDigit(remainder[1]) {
		i = 2
	} else {
		return "", false
	}

	// 3-letter month + 2-digit day
	if i+5 <= len(remainder) && months[remainder[i:i+3]] && isDigit(remainder[i+3]) && isDigit(remainder[i+4]) {
		i += 5
		return remainder[:i], true
	}
	return "", false
}

func extractTeamsFromSubTitle(subTitle string) (string, string, bool) {
	// Try "SAS at OKC" or "OKC at SAS" patterns
	parts := strings.Fields(subTitle)
	if len(parts) >= 3 {
		at := -1
		for i, p := range parts {
			if strings.EqualFold(p, "at") {
				at = i
				break
			}
		}
		if at > 0 && at < len(parts)-1 {
			a := strings.ToUpper(parts[at-1])
			b := strings.ToUpper(parts[at+1])
			if isUpperAlpha(a) && isUpperAlpha(b) {
				return a, b, true
			}
		}
	}
	return "", "", false
}

// teamsMatch checks if two sets of team codes refer to the same matchup (order-insensitive).
func teamsMatch(a1, b1, a2, b2 string) bool {
	if a1 == "" || b1 == "" || a2 == "" || b2 == "" {
		return false
	}
	return (a1 == a2 && b1 == b2) || (a1 == b2 && b1 == a2)
}

// containsTeam checks if a given team code is one of the two game teams.
func containsTeam(gameA, gameB, candidate string) bool {
	if candidate == "" {
		return false
	}
	return candidate == gameA || candidate == gameB
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// isUpperAlpha returns true if the string is 2-4 uppercase letters.
func isUpperAlpha(s string) bool {
	if len(s) < 2 || len(s) > 4 {
		return false
	}
	for _, c := range s {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}

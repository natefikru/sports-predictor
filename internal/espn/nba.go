package espn

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// kalshiToESPN normalizes Kalshi 3-letter codes to ESPN abbreviations.
var kalshiToESPN = map[string]string{
	"SAS": "SA",
	"GSW": "GS",
	"NOP": "NO",
	"UTA": "UTAH",
	"PHX": "PHX",
	"NOR": "NO",
}

func normalizeAbbr(code string) string {
	up := strings.ToUpper(code)
	if mapped, ok := kalshiToESPN[up]; ok {
		return mapped
	}
	return up
}

// FindTeams resolves two team abbreviation strings (Kalshi-style) to TeamInfo
// by fetching the ESPN teams list.
func (c *Client) FindTeams(abbrA, abbrB string) (TeamInfo, TeamInfo, error) {
	espnA := normalizeAbbr(abbrA)
	espnB := normalizeAbbr(abbrB)

	var resp teamsResponse
	if err := c.getJSON(c.siteURL("/apis/site/v2/sports/basketball/nba/teams"), &resp); err != nil {
		return TeamInfo{}, TeamInfo{}, fmt.Errorf("fetch teams: %w", err)
	}

	teams := make(map[string]TeamInfo)
	for _, sp := range resp.Sports {
		for _, lg := range sp.Leagues {
			for _, t := range lg.Teams {
				teams[t.Team.Abbreviation] = TeamInfo{
					ID:           t.Team.ID,
					Abbreviation: t.Team.Abbreviation,
					DisplayName:  t.Team.DisplayName,
					ShortName:    t.Team.ShortDisplayName,
				}
			}
		}
	}

	tA, okA := teams[espnA]
	tB, okB := teams[espnB]
	if !okA {
		return TeamInfo{}, TeamInfo{}, fmt.Errorf("team not found for abbreviation %q (normalized from %q)", espnA, abbrA)
	}
	if !okB {
		return TeamInfo{}, TeamInfo{}, fmt.Errorf("team not found for abbreviation %q (normalized from %q)", espnB, abbrB)
	}
	return tA, tB, nil
}

// FetchSeriesGames returns all games (completed + scheduled) between the two
// teams in the current playoff season, ordered chronologically.
func (c *Client) FetchSeriesGames(teamA, teamB TeamInfo) ([]GameResult, error) {
	year := time.Now().Year()
	url := c.siteURL(fmt.Sprintf("/apis/site/v2/sports/basketball/nba/teams/%s/schedule?season=%d&seasontype=3", teamA.ID, year))

	var resp scheduleResponse
	if err := c.getJSON(url, &resp); err != nil {
		return nil, fmt.Errorf("fetch schedule: %w", err)
	}

	var games []GameResult
	for _, ev := range resp.Events {
		if len(ev.Competitions) == 0 {
			continue
		}
		comp := ev.Competitions[0]

		// Filter to games that involve both teams.
		var home, away struct {
			ID    string
			Abbr  string
			Score string
			Name  string
			Short string
		}
		hasA, hasB := false, false
		for _, comp2 := range comp.Competitors {
			if comp2.Team.Abbreviation == teamA.Abbreviation {
				hasA = true
			}
			if comp2.Team.Abbreviation == teamB.Abbreviation {
				hasB = true
			}
			if comp2.HomeAway == "home" {
				home.ID = comp2.ID
				home.Abbr = comp2.Team.Abbreviation
				home.Score = comp2.Score.DisplayValue
				home.Name = comp2.Team.DisplayName
				home.Short = comp2.Team.ShortDisplayName
			} else {
				away.ID = comp2.ID
				away.Abbr = comp2.Team.Abbreviation
				away.Score = comp2.Score.DisplayValue
				away.Name = comp2.Team.DisplayName
				away.Short = comp2.Team.ShortDisplayName
			}
		}
		if !hasA || !hasB {
			continue
		}

		homeScore, _ := strconv.Atoi(home.Score)
		awayScore, _ := strconv.Atoi(away.Score)

		games = append(games, GameResult{
			GameID: ev.ID,
			Date:   ev.Date,
			HomeTeam: TeamInfo{
				ID:           home.ID,
				Abbreviation: home.Abbr,
				DisplayName:  home.Name,
				ShortName:    home.Short,
			},
			AwayTeam: TeamInfo{
				ID:           away.ID,
				Abbreviation: away.Abbr,
				DisplayName:  away.Name,
				ShortName:    away.Short,
			},
			HomeScore: homeScore,
			AwayScore: awayScore,
			Status:    comp.Status.Type.Name,
		})
	}
	return games, nil
}

// FetchGameRoster returns all players (who did not have DNP) for a team in a game.
func (c *Client) FetchGameRoster(gameID, teamID string) ([]rosterEntry, error) {
	url := c.coreURL(fmt.Sprintf(
		"/v2/sports/basketball/leagues/nba/events/%s/competitions/%s/competitors/%s/roster",
		gameID, gameID, teamID,
	))

	var resp rosterResponse
	if err := c.getJSON(url, &resp); err != nil {
		return nil, fmt.Errorf("fetch roster game=%s team=%s: %w", gameID, teamID, err)
	}

	var active []rosterEntry
	for _, e := range resp.Entries {
		if !e.DidNotPlay && e.Statistics.Ref != "" {
			active = append(active, e)
		}
	}
	return active, nil
}

// FetchPlayerGameStats fetches the per-game box score for one player in one game.
func (c *Client) FetchPlayerGameStats(gameID, teamID, playerID string) (PlayerGameLine, error) {
	url := c.coreURL(fmt.Sprintf(
		"/v2/sports/basketball/leagues/nba/events/%s/competitions/%s/competitors/%s/roster/%s/statistics/0",
		gameID, gameID, teamID, playerID,
	))

	var resp playerStatResponse
	if err := c.getJSON(url, &resp); err != nil {
		return PlayerGameLine{}, err
	}

	line := PlayerGameLine{GameID: gameID}
	for _, cat := range resp.Splits.Categories {
		for _, s := range cat.Stats {
			switch s.Name {
			case "points":
				line.Points = s.Value
			case "rebounds":
				line.Rebounds = s.Value
			case "assists":
				line.Assists = s.Value
			case "steals":
				line.Steals = s.Value
			case "blocks":
				line.Blocks = s.Value
			case "minutes":
				line.Minutes = s.Value
			case "threePointFieldGoalsMade":
				line.ThreePM = s.Value
			case "threePointFieldGoalsAttempted":
				line.ThreePA = s.Value
			case "fieldGoalsMade":
				line.FGM = s.Value
			case "fieldGoalsAttempted":
				line.FGA = s.Value
			case "freeThrowsMade":
				line.FTM = s.Value
			case "freeThrowsAttempted":
				line.FTA = s.Value
			}
		}
	}
	return line, nil
}

// FetchSeriesReport builds the full series report for two teams.
// Only completed games are included in player stat aggregation.
func (c *Client) FetchSeriesReport(abbrA, abbrB string) (*SeriesReport, error) {
	teamA, teamB, err := c.FindTeams(abbrA, abbrB)
	if err != nil {
		return nil, err
	}

	allGames, err := c.FetchSeriesGames(teamA, teamB)
	if err != nil {
		return nil, err
	}

	report := &SeriesReport{
		TeamA: teamA,
		TeamB: teamB,
		Games: allGames,
	}

	// Tally wins for series record display.
	var winsA, winsB int
	var completedGames []GameResult
	for _, g := range allGames {
		if g.Status != "STATUS_FINAL" {
			continue
		}
		completedGames = append(completedGames, g)
		if g.HomeTeam.Abbreviation == teamA.Abbreviation && g.HomeScore > g.AwayScore {
			winsA++
		} else if g.AwayTeam.Abbreviation == teamA.Abbreviation && g.AwayScore > g.HomeScore {
			winsA++
		} else {
			winsB++
		}
		report.AvgTotal += float64(g.HomeScore + g.AwayScore)
	}
	if len(completedGames) > 0 {
		report.AvgTotal /= float64(len(completedGames))
	}
	switch {
	case winsA > winsB:
		report.SeriesRecord = fmt.Sprintf("%s leads %d-%d", teamA.ShortName, winsA, winsB)
	case winsB > winsA:
		report.SeriesRecord = fmt.Sprintf("%s leads %d-%d", teamB.ShortName, winsB, winsA)
	default:
		report.SeriesRecord = fmt.Sprintf("Series tied %d-%d", winsA, winsB)
	}

	// Collect player stats for each completed game.
	// playerMap key = playerID string
	playerMap := make(map[string]*PlayerSeries)

	for _, g := range completedGames {
		// Determine ESPN team IDs for home/away in this specific game.
		homeID := g.HomeTeam.ID
		awayID := g.AwayTeam.ID

		teamPairs := [][2]string{{homeID, g.HomeTeam.Abbreviation}, {awayID, g.AwayTeam.Abbreviation}}
		for _, pair := range teamPairs {
			tid := pair[0]
			abbr := pair[1]
			_ = abbr

			entries, err := c.FetchGameRoster(g.GameID, tid)
			if err != nil {
				fmt.Printf("  [warn] roster fetch failed game=%s team=%s: %v\n", g.GameID, tid, err)
				continue
			}

			for _, entry := range entries {
				pidStr := strconv.Itoa(entry.PlayerID)
				stats, err := c.FetchPlayerGameStats(g.GameID, tid, pidStr)
				if err != nil {
					continue
				}
				// Skip players who barely played (e.g. garbage time only).
				if stats.Minutes < 3 {
					continue
				}

				time.Sleep(60 * time.Millisecond) // be polite

				ps, exists := playerMap[pidStr]
				if !exists {
					ps = &PlayerSeries{
						PlayerID:    pidStr,
						DisplayName: entry.DisplayName,
						TeamID:      tid,
					}
					playerMap[pidStr] = ps
				}
				ps.Lines = append(ps.Lines, stats)
			}
		}
	}

	// Compute averages and collect into report.
	for _, ps := range playerMap {
		n := float64(len(ps.Lines))
		if n == 0 {
			continue
		}
		for _, l := range ps.Lines {
			ps.Avg.Points += l.Points
			ps.Avg.Rebounds += l.Rebounds
			ps.Avg.Assists += l.Assists
			ps.Avg.Steals += l.Steals
			ps.Avg.Blocks += l.Blocks
			ps.Avg.Minutes += l.Minutes
			ps.Avg.ThreePM += l.ThreePM
			ps.Avg.FGM += l.FGM
			ps.Avg.FGA += l.FGA
		}
		ps.Avg.Points /= n
		ps.Avg.Rebounds /= n
		ps.Avg.Assists /= n
		ps.Avg.Steals /= n
		ps.Avg.Blocks /= n
		ps.Avg.Minutes /= n
		ps.Avg.ThreePM /= n
		ps.Avg.FGM /= n
		ps.Avg.FGA /= n

		report.Players = append(report.Players, ps)
	}

	return report, nil
}

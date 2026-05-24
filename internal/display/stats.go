package display

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/natefikru/sports-predictor/internal/espn"
)

func PrintSeriesReport(r *espn.SeriesReport) {
	width := 62
	border := strings.Repeat("═", width)
	fmt.Println(styleHeader.Render(border))
	fmt.Printf("  %s vs %s\n", styleBold.Render(r.TeamA.DisplayName), styleBold.Render(r.TeamB.DisplayName))
	fmt.Printf("  %s\n", styleDim.Render(r.SeriesRecord))
	if r.AvgTotal > 0 {
		fmt.Printf("  %s\n", styleDim.Render(fmt.Sprintf("Series avg total: %.1f pts", r.AvgTotal)))
	}
	fmt.Println(styleHeader.Render(border))
	fmt.Println()

	// Game-by-game results
	fmt.Println(styleSectionHd.Render("GAME-BY-GAME RESULTS"))
	fmt.Printf("  %-8s  %-10s  %-32s  %s\n",
		styleColHd.Render("Game"),
		styleColHd.Render("Date"),
		styleColHd.Render("Matchup"),
		styleColHd.Render("Score"),
	)

	gameNum := 0
	for _, g := range r.Games {
		gameNum++
		dateStr := fmtESPNDate(g.Date)
		if g.Status == "STATUS_FINAL" {
			matchup := fmt.Sprintf("%s at %s", g.AwayTeam.ShortName, g.HomeTeam.ShortName)
			score := fmt.Sprintf("%d - %d", g.AwayScore, g.HomeScore)
			var winner string
			if g.HomeScore > g.AwayScore {
				winner = g.HomeTeam.ShortName + " wins"
			} else {
				winner = g.AwayTeam.ShortName + " wins"
			}
			fmt.Printf("  Game %-3d  %-10s  %-32s  %s %s\n",
				gameNum, dateStr, matchup, score, styleDim.Render("("+winner+")"))
		} else {
			matchup := fmt.Sprintf("%s at %s", g.AwayTeam.ShortName, g.HomeTeam.ShortName)
			fmt.Printf("  Game %-3d  %-10s  %-32s  %s\n",
				gameNum, dateStr, matchup, styleDim.Render("(scheduled)"))
		}
	}
	fmt.Println()

	// Separate players by team.
	teamPlayers := make(map[string][]*espn.PlayerSeries)
	for _, ps := range r.Players {
		teamPlayers[ps.TeamID] = append(teamPlayers[ps.TeamID], ps)
	}

	// Sort teams: TeamA first, then TeamB.
	orderedTeamIDs := []string{r.TeamA.ID, r.TeamB.ID}

	// Completed game count (for per-game column headers).
	completedCount := 0
	for _, g := range r.Games {
		if g.Status == "STATUS_FINAL" {
			completedCount++
		}
	}

	for _, tid := range orderedTeamIDs {
		players, ok := teamPlayers[tid]
		if !ok {
			continue
		}

		var teamName string
		if tid == r.TeamA.ID {
			teamName = r.TeamA.DisplayName
		} else {
			teamName = r.TeamB.DisplayName
		}

		// Sort by avg minutes descending.
		sort.Slice(players, func(i, j int) bool {
			return players[i].Avg.Minutes > players[j].Avg.Minutes
		})

		fmt.Println(styleSectionHd.Render(strings.ToUpper(teamName) + " — PLAYER STATS"))

		// Header row: name + avg cols + per-game cols
		header := fmt.Sprintf("  %-22s  %5s  %5s  %5s  %5s  %5s  %4s",
			styleColHd.Render("Player"),
			styleColHd.Render("MIN"),
			styleColHd.Render("PTS"),
			styleColHd.Render("REB"),
			styleColHd.Render("AST"),
			styleColHd.Render("BLK/STL"),
			styleColHd.Render("FG%"),
		)
		// Add per-game columns
		for i := 1; i <= completedCount; i++ {
			header += fmt.Sprintf("  %s", styleColHd.Render(fmt.Sprintf("G%d", i)))
		}
		fmt.Println(header)

		for _, ps := range players {
			fgPct := 0.0
			if ps.Avg.FGA > 0 {
				fgPct = ps.Avg.FGM / ps.Avg.FGA * 100
			}
			row := fmt.Sprintf("  %-22s  %5.1f  %5.1f  %5.1f  %5.1f  %5s  %4.1f%%",
				truncate(ps.DisplayName, 22),
				ps.Avg.Minutes,
				ps.Avg.Points,
				ps.Avg.Rebounds,
				ps.Avg.Assists,
				fmt.Sprintf("%.1f/%.1f", ps.Avg.Blocks, ps.Avg.Steals),
				fgPct,
			)
			// Per-game PTS
			for i, line := range ps.Lines {
				if i >= completedCount {
					break
				}
				row += fmt.Sprintf("  %s", colorPts(line.Points))
			}
			fmt.Println(row)
		}
		fmt.Println()
	}
}

func fmtESPNDate(dateStr string) string {
	// ESPN dates come as "2026-05-19T00:30Z" (no seconds)
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04Z", "2006-01-02T15:04:05Z", "2006-01-02"} {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t.Format("Jan 2")
		}
	}
	if len(dateStr) >= 10 {
		return dateStr[:10]
	}
	return dateStr
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func colorPts(pts float64) string {
	s := fmt.Sprintf("%4.0f", pts)
	switch {
	case pts >= 25:
		return styleGreen.Render(s)
	case pts >= 15:
		return styleYellow.Render(s)
	default:
		return styleDim.Render(s)
	}
}

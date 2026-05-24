package display

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/natefikru/sports-predictor/internal/kalshi"
)

var (
	styleBold      = lipgloss.NewStyle().Bold(true)
	styleHeader    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	styleDivider   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleSectionHd = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	styleDim       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleGreen     = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	styleYellow    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	styleRed       = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleColHd     = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
)

// PickGame presents a numbered list of game events and returns the one the user selects.
func PickGame(events []kalshi.Event) (kalshi.Event, error) {
	if len(events) == 0 {
		return kalshi.Event{}, fmt.Errorf("no open NBA games found on Kalshi right now")
	}
	if len(events) == 1 {
		fmt.Printf("Only one game available: %s\n\n", events[0].Title)
		return events[0], nil
	}

	fmt.Println(styleHeader.Render("Available NBA Games:"))
	for i, e := range events {
		subtitle := fmtSubtitle(e.SubTitle)
		fmt.Printf("  [%d] %-50s %s\n", i+1, e.Title, styleDim.Render(subtitle))
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Select game [1-%d]: ", len(events))
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		n, err := strconv.Atoi(line)
		if err == nil && n >= 1 && n <= len(events) {
			fmt.Println()
			return events[n-1], nil
		}
		fmt.Printf("  Please enter a number between 1 and %d.\n", len(events))
	}
}

// PrintGameView renders the complete game view with all market sections.
func PrintGameView(view kalshi.GameView) {
	printGameHeader(view.GameEvent)

	// Split sections into three top-level groups for visual separation
	var gameMarkets, seriesMarkets, propMarkets []kalshi.MarketSection
	propScopes := map[kalshi.ScopeGroup]bool{
		kalshi.ScopePlayerPoints:       true,
		kalshi.ScopePlayer3pt:          true,
		kalshi.ScopePlayerRebounds:     true,
		kalshi.ScopePlayerAssists:      true,
		kalshi.ScopePlayerSteals:       true,
		kalshi.ScopePlayerCombined:     true,
		kalshi.ScopePlayerDoubleDouble: true,
	}
	seriesScopes := map[kalshi.ScopeGroup]bool{
		kalshi.ScopeSeriesWinner:  true,
		kalshi.ScopeSeriesSpread:  true,
		kalshi.ScopeSeriesGames:   true,
		kalshi.ScopeSeriesExact:   true,
		kalshi.ScopeSeriesGame7:   true,
	}

	for _, s := range view.Sections {
		switch {
		case propScopes[s.Scope]:
			propMarkets = append(propMarkets, s)
		case seriesScopes[s.Scope]:
			seriesMarkets = append(seriesMarkets, s)
		default:
			gameMarkets = append(gameMarkets, s)
		}
	}

	fmt.Println(styleDivider.Render("─── GAME MARKETS " + strings.Repeat("─", 44)))
	for _, s := range gameMarkets {
		printSection(s)
	}

	fmt.Println()
	fmt.Println(styleDivider.Render("─── SERIES MARKETS " + strings.Repeat("─", 42)))
	for _, s := range seriesMarkets {
		printSection(s)
	}

	hasPropMarkets := false
	for _, s := range propMarkets {
		if len(s.Markets) > 0 {
			hasPropMarkets = true
			break
		}
	}
	if hasPropMarkets {
		fmt.Println()
		fmt.Println(styleDivider.Render("─── PLAYER PROPS " + strings.Repeat("─", 44)))
		for _, s := range propMarkets {
			printSection(s)
		}
	}
}

func printGameHeader(e kalshi.Event) {
	width := 62
	border := strings.Repeat("═", width)
	fmt.Println(styleHeader.Render(border))
	fmt.Printf("  %s\n", styleBold.Render(e.Title))
	if e.SubTitle != "" {
		fmt.Printf("  %s\n", styleDim.Render(e.SubTitle))
	}
	fmt.Println(styleHeader.Render(border))
	fmt.Println()
}

func printSection(s kalshi.MarketSection) {
	fmt.Println()
	fmt.Println(styleSectionHd.Render(s.Label))

	if s.FetchErr {
		fmt.Println(styleRed.Render("  (fetch error)"))
		return
	}
	if len(s.Markets) == 0 {
		fmt.Println(styleDim.Render("  (no open markets)"))
		return
	}

	// Determine if this is a moneyline-style section (two team outcomes)
	// vs a spread/total/prop section (multiple lines with yes_sub_title as label)
	isMoneyline := s.Scope == kalshi.ScopeMoneyline || s.Scope == kalshi.ScopeSeriesWinner

	if isMoneyline {
		printMoneylineTable(s)
	} else {
		printLinesTable(s)
	}
}

func printMoneylineTable(s kalshi.MarketSection) {
	fmt.Printf("  %-28s %-8s %-8s %s\n",
		styleColHd.Render("Outcome"),
		styleColHd.Render("Bid"),
		styleColHd.Render("Ask"),
		styleColHd.Render("Implied"),
	)

	var totalVol, totalOI float64
	for _, m := range s.Markets {
		bid := parseFloat(m.YesBid)
		ask := parseFloat(m.YesAsk)
		implied := (bid + ask) / 2.0 * 100

		label := m.YesSubTitle
		if label == "" {
			label = m.Title
		}
		// Truncate long labels
		if len(label) > 26 {
			label = label[:23] + "..."
		}

		impliedStr := fmt.Sprintf("%.1f%%", implied)
		fmt.Printf("  %-28s %-8s %-8s %s\n",
			label,
			fmt.Sprintf("$%.2f", bid),
			fmt.Sprintf("$%.2f", ask),
			colorImplied(implied, impliedStr),
		)
		totalVol += parseFloat(m.Volume24hFP)
		totalOI += parseFloat(m.OpenInterestFP)
	}

	if totalVol > 0 || totalOI > 0 {
		fmt.Printf("  %s\n", styleDim.Render(
			fmt.Sprintf("Volume 24h: $%s  |  Open Interest: $%s",
				fmtDollar(totalVol), fmtDollar(totalOI)),
		))
	}
}

func printLinesTable(s kalshi.MarketSection) {
	fmt.Printf("  %-32s %-8s %-8s %s\n",
		styleColHd.Render("Line"),
		styleColHd.Render("Bid"),
		styleColHd.Render("Ask"),
		styleColHd.Render("Implied"),
	)

	for _, m := range s.Markets {
		bid := parseFloat(m.YesBid)
		ask := parseFloat(m.YesAsk)
		implied := (bid + ask) / 2.0 * 100

		label := m.YesSubTitle
		if label == "" {
			label = m.Title
		}
		if len(label) > 30 {
			label = label[:27] + "..."
		}

		impliedStr := fmt.Sprintf("%.1f%%", implied)
		fmt.Printf("  %-32s %-8s %-8s %s\n",
			label,
			fmt.Sprintf("$%.2f", bid),
			fmt.Sprintf("$%.2f", ask),
			colorImplied(implied, impliedStr),
		)
	}
}

func colorImplied(pct float64, s string) string {
	switch {
	case pct > 60:
		return styleGreen.Render(s)
	case pct >= 40:
		return styleYellow.Render(s)
	default:
		return styleRed.Render(s)
	}
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

func fmtDollar(v float64) string {
	if v >= 1_000_000 {
		return fmt.Sprintf("%.1fM", v/1_000_000)
	}
	if v >= 1_000 {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.2f", v)
}

func fmtSubtitle(s string) string {
	if s == "" {
		return ""
	}
	// Try to extract a date part from sub_title like "NYK at CLE (May 25)"
	if idx := strings.Index(s, "("); idx >= 0 {
		inner := strings.TrimSuffix(s[idx+1:], ")")
		// Parse "May 25" → "May 25"
		if t, err := time.Parse("Jan 2", inner); err == nil {
			return t.Format("Jan 2")
		}
		return inner
	}
	return s
}

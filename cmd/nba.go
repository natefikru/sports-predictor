package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/natefikru/sports-predicter/internal/display"
	"github.com/natefikru/sports-predicter/internal/kalshi"
)

var gameFlag string

var nbaCmd = &cobra.Command{
	Use:   "nba",
	Short: "View NBA betting markets from Kalshi",
	Long:  "Fetch and display all open NBA market lines (moneyline, spreads, totals, series, props) from Kalshi.",
	RunE:  runNBA,
}

func init() {
	nbaCmd.Flags().StringVar(&gameFlag, "game", "", `Filter to a specific game by team codes, e.g. "OKC SAS"`)
}

func runNBA(cmd *cobra.Command, args []string) error {
	client := kalshi.NewClient()

	// Step 1: Discover and classify all NBA series
	fmt.Fprintln(os.Stderr, "Discovering NBA series...")
	allScopes, err := kalshi.DiscoverAndClassifyNBASeries(client)
	if err != nil {
		return fmt.Errorf("series discovery: %w", err)
	}

	// Step 2: Find the moneyline series ticker for the game picker
	var moneylineTicker string
	for _, s := range allScopes {
		if s.Group == kalshi.ScopeMoneyline {
			moneylineTicker = s.Ticker
			break
		}
	}
	if moneylineTicker == "" {
		return fmt.Errorf("no moneyline series found in discovered NBA series")
	}

	// Step 3: Fetch game events
	fmt.Fprintln(os.Stderr, "Fetching NBA games...")
	gameEvents, err := kalshi.FetchGameEvents(client, moneylineTicker)
	if err != nil {
		return fmt.Errorf("fetching games: %w", err)
	}

	// Step 4: Apply --game filter if provided
	if gameFlag != "" {
		gameEvents = filterByTeams(gameEvents, gameFlag)
		if len(gameEvents) == 0 {
			return fmt.Errorf("no games found matching %q", gameFlag)
		}
	}

	// Step 5: User picks a game
	selected, err := display.PickGame(gameEvents)
	if err != nil {
		return err
	}

	// Step 6: Build full game view (concurrent fetches for all scopes)
	fmt.Fprintln(os.Stderr, "Fetching all markets...")
	view, err := kalshi.BuildGameView(client, selected, allScopes)
	if err != nil {
		return fmt.Errorf("building game view: %w", err)
	}

	// Step 7: Display
	display.PrintGameView(view)
	return nil
}

// filterByTeams filters game events to those containing both team codes from the flag.
func filterByTeams(events []kalshi.Event, flag string) []kalshi.Event {
	codes := strings.Fields(strings.ToUpper(flag))
	if len(codes) < 1 {
		return events
	}

	var filtered []kalshi.Event
	for _, e := range events {
		ticker := strings.ToUpper(e.EventTicker)
		match := true
		for _, code := range codes {
			if !strings.Contains(ticker, code) {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

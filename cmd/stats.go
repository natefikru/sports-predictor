package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/natefikru/sports-predictor/internal/display"
	"github.com/natefikru/sports-predictor/internal/espn"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Fetch real NBA player stats for a series matchup",
	Long:  "Pull per-player per-game stats from ESPN for a playoff series. Used to cross-reference Kalshi market odds with actual performance data.",
	RunE:  runStats,
}

var statsGameFlag string

func init() {
	statsCmd.Flags().StringVar(&statsGameFlag, "game", "", `Two team codes to analyze, e.g. "OKC SAS"`)
	rootCmd.AddCommand(statsCmd)
}

func runStats(cmd *cobra.Command, args []string) error {
	if statsGameFlag == "" {
		return fmt.Errorf("--game is required (e.g. --game \"OKC SAS\")")
	}

	codes := strings.Fields(strings.ToUpper(statsGameFlag))
	if len(codes) < 2 {
		return fmt.Errorf("--game must contain two team codes (e.g. \"OKC SAS\")")
	}

	client := espn.NewClient()

	fmt.Fprintln(os.Stderr, "Looking up teams...")
	fmt.Fprintln(os.Stderr, "Fetching series schedule and player stats (this takes ~20–40s)...")

	report, err := client.FetchSeriesReport(codes[0], codes[1])
	if err != nil {
		return fmt.Errorf("fetch series report: %w", err)
	}

	display.PrintSeriesReport(report)
	return nil
}

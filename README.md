# sports-predictor

A Go CLI tool for NBA betting analysis. Pulls live market odds from [Kalshi](https://kalshi.com) and real per-game player stats from ESPN, then cross-references them to surface mispriced lines.

## Install

```bash
git clone https://github.com/natefikru/sports-predictor.git
cd sports-predictor
go build -o sp .
```

## Commands

### `sp nba` — Kalshi market viewer

Browse all open NBA betting markets for a game.

```bash
# Pick from today's games interactively
./sp nba

# Jump straight to a specific matchup
./sp nba --game "OKC SAS"
```

**Output sections:**

- **GAME MARKETS** — moneyline, spread, total, 1st/2nd half lines, quarter markets
- **SERIES MARKETS** — series winner, spread, total games, exact score, game 7
- **PLAYER PROPS** — points, 3-pointers, rebounds, assists, steals, combined, double-doubles

Each line shows bid, ask, and implied probability (color-coded: green >60%, yellow 40–60%, red <40%).

Kalshi markets are discovered dynamically via `GET /series?tags=Basketball` — no hardcoded tickers, no API key required.

---

### `sp stats` — ESPN player stats

Pull per-player per-game box scores for a playoff series and display them in a sortable table.

```bash
./sp stats --game "OKC SAS"
```

**Output:**
- Series record and game-by-game scores
- Both rosters ranked by minutes, with series averages: MIN, PTS, REB, AST, BLK/STL, FG%
- Per-game point totals color-coded across all completed games

Stats are fetched from ESPN's API (no key required). Each completed game's full roster and individual box scores are fetched and aggregated.

---

## Example workflow

```bash
# 1. See what Kalshi is pricing for player assists tonight
./sp nba --game "OKC SAS"

# 2. Check what players are actually producing in this series
./sp stats --game "OKC SAS"

# 3. Cross-reference: find lines where Kalshi's implied probability
#    is far below the player's actual series rate
```

The core edge: Kalshi prices props using season-long averages. Series-specific data (pace, matchup, role changes) often tells a different story.

## Dependencies

- [`cobra`](https://github.com/spf13/cobra) — CLI framework
- [`lipgloss`](https://github.com/charmbracelet/lipgloss) — terminal styling

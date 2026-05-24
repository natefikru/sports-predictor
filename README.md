# sports-predictor

A Go CLI tool that pulls live NBA betting markets from [Kalshi](https://kalshi.com) and displays them in a clean terminal view — moneyline, spreads, totals, series markets, and player props.

## Install

```bash
git clone https://github.com/natefikru/sports-predictor.git
cd sports-predictor
go build -o sp .
```

## Usage

```bash
# Browse today's NBA games and pick one interactively
./sp nba

# Jump straight to a specific matchup
./sp nba --game "OKC SAS"
```

## Output

Markets are grouped into three sections:

**GAME MARKETS** — moneyline, game spread, game total, 1st/2nd half lines, quarter markets

**SERIES MARKETS** — series winner, series game spread, total games, exact score, goes to game 7

**PLAYER PROPS** — points, 3-pointers, rebounds, assists, steals, combined stats, double-doubles

Each line shows bid, ask, and implied probability (color-coded: green >60%, yellow 40–60%, red <40%).

## How it works

Series are discovered dynamically via `GET /series?tags=Basketball` — no hardcoded tickers. All market data comes from Kalshi's public REST API; no API key required.

## Dependencies

- [`cobra`](https://github.com/spf13/cobra) — CLI framework
- [`lipgloss`](https://github.com/charmbracelet/lipgloss) — terminal styling

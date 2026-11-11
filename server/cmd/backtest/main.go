package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"server/internal/logic/football"
)

func main() {
	league := flag.String("league", "E0", "football-data league code, for example E0")
	fromYear := flag.Int("from-year", 2020, "first calendar year to include in evaluation")
	flag.Parse()

	report, err := football.BacktestLeague(context.Background(), *league, *fromYear)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

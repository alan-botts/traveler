package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"traveler/internal/google"
	"traveler/internal/tui"
)

var flightsCmd = &cobra.Command{
	Use:   "flights <origin> <destination> <date>",
	Short: "Search for one-way flights",
	Long:  "Search Google Flights for one-way flights. Date format: YYYY-MM-DD.",
	Args:  cobra.ExactArgs(3),
	RunE:  runFlights,
}

func init() {
	rootCmd.AddCommand(flightsCmd)
}

func runFlights(cmd *cobra.Command, args []string) error {
	origin := strings.ToUpper(args[0])
	destination := strings.ToUpper(args[1])
	date := args[2]

	// Validate airport codes (3 letters).
	airportRe := regexp.MustCompile(`^[A-Z]{3}$`)
	if !airportRe.MatchString(origin) {
		return fmt.Errorf("invalid origin airport code: %s (expected 3-letter IATA code)", origin)
	}
	if !airportRe.MatchString(destination) {
		return fmt.Errorf("invalid destination airport code: %s (expected 3-letter IATA code)", destination)
	}

	// Validate date format.
	dateRe := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRe.MatchString(date) {
		return fmt.Errorf("invalid date format: %s (expected YYYY-MM-DD)", date)
	}

	fmt.Printf("Searching flights: %s -> %s on %s...\n", origin, destination, date)

	client := google.NewClient()

	flights, err := client.SearchFlights(origin, destination, date)
	if err != nil {
		return fmt.Errorf("flight search failed: %w", err)
	}

	if len(flights) == 0 {
		fmt.Println("No flights found.")
		return nil
	}

	// Launch TUI.
	model := tui.NewModel(flights, origin, destination, date)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		return err
	}

	return nil
}

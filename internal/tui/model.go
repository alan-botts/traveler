package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"traveler/internal/google"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	rowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("15"))

	bestBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// Model is the Bubbletea model for displaying flight search results.
type Model struct {
	flights     []google.Flight
	cursor      int
	origin      string
	destination string
	date        string
	width       int
	height      int
}

// NewModel creates a new TUI model with the given flight results.
func NewModel(flights []google.Flight, origin, destination, date string) Model {
	return Model{
		flights:     flights,
		origin:      origin,
		destination: destination,
		date:        date,
		width:       120,
		height:      40,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.flights)-1 {
				m.cursor++
			}
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			m.cursor = len(m.flights) - 1
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	title := titleStyle.Render(fmt.Sprintf("Flights: %s -> %s on %s (%d results)",
		m.origin, m.destination, m.date, len(m.flights)))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Column headers.
	header := fmt.Sprintf("%-6s %-10s %-6s %-6s %-8s %-8s %-8s %-7s %s",
		"Type", "Airline", "Flight", "Dep", "Arr", "Dur", "Stops", "Price", "Route")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Determine visible window.
	visibleRows := m.height - 7
	if visibleRows < 5 {
		visibleRows = 5
	}
	start := 0
	if m.cursor >= visibleRows {
		start = m.cursor - visibleRows + 1
	}
	end := start + visibleRows
	if end > len(m.flights) {
		end = len(m.flights)
	}

	for i := start; i < end; i++ {
		f := m.flights[i]
		row := formatFlightRow(f)

		style := rowStyle
		if i == m.cursor {
			style = selectedStyle
		}
		b.WriteString(style.Render(row))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("j/k or arrows to navigate | q to quit"))

	return b.String()
}

func formatFlightRow(f google.Flight) string {
	category := "other"
	if f.Category == "best" {
		category = bestBadge.Render("BEST")
	}

	// Primary leg info (first leg).
	airline := ""
	flightNum := ""
	depTime := ""
	arrTime := ""
	route := ""

	if len(f.Legs) > 0 {
		first := f.Legs[0]
		last := f.Legs[len(f.Legs)-1]
		airline = first.AirlineCode
		flightNum = first.FlightNum
		depTime = fmt.Sprintf("%02d:%02d", first.DepTime[0], first.DepTime[1])
		arrTime = fmt.Sprintf("%02d:%02d", last.ArrTime[0], last.ArrTime[1])
		route = fmt.Sprintf("%s->%s", first.DepAirport, last.ArrAirport)

		// Show connections.
		if len(f.Legs) > 1 {
			var via []string
			for i := 0; i < len(f.Legs)-1; i++ {
				via = append(via, f.Legs[i].ArrAirport)
			}
			route += " via " + strings.Join(via, ",")
		}
	}

	stops := len(f.Legs) - 1
	stopsStr := "nonstop"
	if stops == 1 {
		stopsStr = "1 stop"
	} else if stops > 1 {
		stopsStr = fmt.Sprintf("%d stops", stops)
	}

	durStr := formatDuration(f.TotalDuration)
	priceStr := fmt.Sprintf("$%.0f", f.Price)

	return fmt.Sprintf("%-6s %-10s %-6s %-6s %-8s %-8s %-8s %-7s %s",
		category, airline, flightNum, depTime, arrTime, durStr, stopsStr, priceStr, route)
}

func formatDuration(minutes int) string {
	h := minutes / 60
	m := minutes % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

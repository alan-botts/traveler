package google

// Flight represents a complete flight itinerary (possibly with connections).
type Flight struct {
	Legs          []Leg
	TotalDuration int     // total trip duration in minutes
	Price         float64 // price in USD
	Category      string  // "best" or "other"
}

// Leg represents a single flight leg (one takeoff and landing).
type Leg struct {
	AirlineCode string
	FlightNum   string
	DepAirport  string
	ArrAirport  string
	DepDate     [3]int // [year, month, day]
	DepTime     [2]int // [hour, minute]
	ArrDate     [3]int // [year, month, day]
	ArrTime     [2]int // [hour, minute]
	Duration    int    // minutes
}

package google

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseResponse takes the raw Google Flights API response body and extracts
// flight results from it.
//
// Decoding chain:
//  1. Strip XSSI prefix ")]}'\n"
//  2. JSON parse
//  3. Navigate to result[0][2] — inner JSON string
//  4. JSON parse that inner string
//  5. Best flights at [2][0], other flights at [3][0]
func ParseResponse(body []byte) ([]Flight, error) {
	text := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(body)), ")]}'"))
	var outer []interface{}
	if err := json.Unmarshal([]byte(text), &outer); err != nil {
		return nil, fmt.Errorf("parse outer JSON: %w", err)
	}
	for _, value := range outer {
		frame, ok := value.([]interface{})
		if !ok || len(frame) < 3 || frame[0] != "wrb.fr" {
			continue
		}
		payload, ok := frame[2].(string)
		if !ok {
			if len(frame) > 5 {
				return nil, fmt.Errorf("Google Flights RPC error: %v", frame[5])
			}
			return nil, fmt.Errorf("Google Flights RPC missing result payload")
		}
		var inner []interface{}
		if err := json.Unmarshal([]byte(payload), &inner); err != nil {
			return nil, fmt.Errorf("parse inner JSON: %w", err)
		}
		return parseResults(inner)
	}
	return nil, fmt.Errorf("Google Flights response missing result frame")
}

func parseResults(inner []interface{}) ([]Flight, error) {
	if len(inner) < 4 {
		return nil, fmt.Errorf("Google Flights result is incomplete")
	}

	var flights []Flight

	// Step 5a: Best flights at inner[2][0].
	bestFlights := extractFlightGroup(inner, 2, "best")
	flights = append(flights, bestFlights...)

	// Step 5b: Other flights at inner[3][0].
	otherFlights := extractFlightGroup(inner, 3, "other")
	flights = append(flights, otherFlights...)

	return flights, nil
}

func extractFlightGroup(inner []interface{}, groupIdx int, category string) []Flight {
	if groupIdx >= len(inner) || inner[groupIdx] == nil {
		return nil
	}
	group, ok := inner[groupIdx].([]interface{})
	if !ok || len(group) == 0 || group[0] == nil {
		return nil
	}
	items, ok := group[0].([]interface{})
	if !ok {
		return nil
	}

	var flights []Flight
	for _, item := range items {
		f, err := parseFlight(item, category)
		if err != nil {
			continue // skip unparseable flights
		}
		flights = append(flights, f)
	}
	return flights
}

func parseFlight(data interface{}, category string) (Flight, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return Flight{}, fmt.Errorf("flight data is not an array")
	}

	flight := Flight{Category: category}

	// Total duration: data[0][9]
	if flightData, ok := safeIndex(arr, 0); ok {
		if fd, ok := flightData.([]interface{}); ok {
			if dur, ok := safeIndex(fd, 9); ok {
				flight.TotalDuration = toInt(dur)
			}
			// Legs: data[0][2]
			if legsRaw, ok := safeIndex(fd, 2); ok {
				if legs, ok := legsRaw.([]interface{}); ok {
					for _, legRaw := range legs {
						leg, err := parseLeg(legRaw)
						if err != nil {
							return Flight{}, err
						}
						flight.Legs = append(flight.Legs, leg)
					}
				}
			}
		}
	}

	// Price: data[1][0][-1] (last element of data[1][0])
	if priceArr, ok := safeIndex(arr, 1); ok {
		if pa, ok := priceArr.([]interface{}); ok {
			if inner, ok := safeIndex(pa, 0); ok {
				if ia, ok := inner.([]interface{}); ok && len(ia) > 0 {
					flight.Price = toFloat(ia[len(ia)-1])
				}
			}
		}
	}

	if len(flight.Legs) == 0 || flight.TotalDuration <= 0 || flight.Price <= 0 {
		return Flight{}, fmt.Errorf("incomplete or unpriced itinerary")
	}
	return flight, nil
}

func parseLeg(data interface{}) (Leg, error) {
	fl, ok := data.([]interface{})
	if !ok {
		return Leg{}, fmt.Errorf("leg data is not an array")
	}

	leg := Leg{}

	// Airline code: fl[22][0], flight number: fl[22][1]
	if airlineInfo, ok := safeIndex(fl, 22); ok {
		if ai, ok := airlineInfo.([]interface{}); ok {
			if len(ai) > 0 {
				leg.AirlineCode = toString(ai[0])
			}
			if len(ai) > 1 {
				leg.FlightNum = toString(ai[1])
			}
		}
	}

	// Departure airport: fl[3]
	if v, ok := safeIndex(fl, 3); ok {
		leg.DepAirport = toString(v)
	}

	// Arrival airport: fl[6]
	if v, ok := safeIndex(fl, 6); ok {
		leg.ArrAirport = toString(v)
	}

	// Departure date: fl[20] -> [y, m, d]
	if v, ok := safeIndex(fl, 20); ok {
		leg.DepDate = toIntTriple(v)
	}

	// Departure time: fl[8] -> [h, m]
	if v, ok := safeIndex(fl, 8); ok {
		leg.DepTime = toIntPair(v)
	}

	// Arrival date: fl[21] -> [y, m, d]
	if v, ok := safeIndex(fl, 21); ok {
		leg.ArrDate = toIntTriple(v)
	}

	// Arrival time: fl[10] -> [h, m]
	if v, ok := safeIndex(fl, 10); ok {
		leg.ArrTime = toIntPair(v)
	}

	// Duration: fl[11]
	if v, ok := safeIndex(fl, 11); ok {
		leg.Duration = toInt(v)
	}

	if len(leg.DepAirport) != 3 || len(leg.ArrAirport) != 3 || leg.DepDate[0] == 0 || leg.ArrDate[0] == 0 || leg.Duration <= 0 {
		return Leg{}, fmt.Errorf("incomplete flight leg")
	}
	return leg, nil
}

// Helper functions for safe access into untyped JSON arrays.

func safeIndex(arr []interface{}, i int) (interface{}, bool) {
	if i < len(arr) && arr[i] != nil {
		return arr[i], true
	}
	return nil, false
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case json.Number:
		f, _ := n.Float64()
		return f
	}
	return 0
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func toIntTriple(v interface{}) [3]int {
	arr, ok := v.([]interface{})
	if !ok || len(arr) < 3 {
		return [3]int{}
	}
	return [3]int{toInt(arr[0]), toInt(arr[1]), toInt(arr[2])}
}

func toIntPair(v interface{}) [2]int {
	arr, ok := v.([]interface{})
	if !ok || len(arr) == 0 {
		return [2]int{}
	}
	result := [2]int{toInt(arr[0]), 0}
	if len(arr) > 1 {
		result[1] = toInt(arr[1])
	}
	return result
}

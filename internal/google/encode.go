package google

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// BuildRequestBody encodes a flight search request into the form body expected
// by the Google Flights internal API.
//
// Encoding chain:
//  1. Build nested array structure
//  2. JSON stringify (compact)
//  3. Wrap in [null, "<json_string>"]
//  4. JSON stringify that wrapper
//  5. URL encode
//  6. Prepend "f.req="
func BuildRequestBody(origin, destination, date string) (string, error) {
	// Build the flight segment.
	segment := []interface{}{
		[]interface{}{[]interface{}{[]interface{}{origin, 0}}},   // departure airport
		[]interface{}{[]interface{}{[]interface{}{destination, 0}}}, // arrival airport
		nil, // time restrictions
		0,   // max_stops: 0 = ANY
		nil, // airline filter
		nil,
		date, // YYYY-MM-DD
		nil,  // max duration
		nil,  // selected flights
		nil,  // layover airports
		nil, nil, nil, nil, 3,
	}

	// Build the inner request structure.
	inner := []interface{}{
		[]interface{}{},
		[]interface{}{
			nil, nil,
			2,                      // trip_type: 2 = ONE_WAY
			nil, []interface{}{},
			1,                      // seat_type: 1 = ECONOMY
			[]interface{}{1, 0, 0, 0}, // passengers: 1 adult
			nil,                    // price limit
			nil, nil, nil, nil, nil,
			[]interface{}{segment}, // flight segments
			nil, nil, nil, 1,
		},
		2, // sort_by: 2 = CHEAPEST
		0, 0, 2,
	}

	// Step 1-2: JSON stringify the inner structure.
	innerJSON, err := json.Marshal(inner)
	if err != nil {
		return "", fmt.Errorf("marshal inner: %w", err)
	}

	// Step 3-4: Wrap in [null, "<json_string>"] and stringify.
	wrapper := []interface{}{nil, string(innerJSON)}
	wrapperJSON, err := json.Marshal(wrapper)
	if err != nil {
		return "", fmt.Errorf("marshal wrapper: %w", err)
	}

	// Step 5-6: URL encode and prepend f.req=.
	encoded := url.QueryEscape(string(wrapperJSON))
	return "f.req=" + encoded, nil
}

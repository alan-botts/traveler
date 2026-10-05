package google

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	http "github.com/bogdanfinn/fhttp"
)

// The public search page embeds the same result arrays as the RPC. Use it
// when the undocumented RPC rejects a request; never interpret an error as
// an empty search. Pin the market/currency to match our USD output.
func (c *Client) searchPage(origin, destination, date string) ([]Flight, error) {
	c.rateLimit()
	query := url.Values{"hl": {"en"}, "gl": {"US"}, "curr": {"USD"},
		"q": {fmt.Sprintf("one way flights from %s to %s on %s", origin, destination, date)}}
	req, err := http.NewRequest("GET", "https://www.google.com/travel/flights?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	return parsePage(body, origin, destination, date)
}

func pageData(body []byte, key string) ([]interface{}, error) {
	pattern := regexp.MustCompile(`AF_initDataCallback\(\{key:\s*'` + regexp.QuoteMeta(key) + `'[^\n]*?data:`)
	loc := pattern.FindIndex(body)
	if loc == nil {
		return nil, fmt.Errorf("search page missing %s data (consent, challenge, or changed format)", key)
	}
	var data []interface{}
	if err := json.NewDecoder(strings.NewReader(string(body[loc[1]:]))).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode %s: %w", key, err)
	}
	return data, nil
}

func at(value interface{}, path ...int) interface{} {
	for _, i := range path {
		list, ok := value.([]interface{})
		if !ok || i < 0 || i >= len(list) {
			return nil
		}
		value = list[i]
	}
	return value
}

// Google sometimes resolves BER to the Berlin city rather than the airport.
// Accept only that known city entity, then verify the actual flight endpoints.
func pageAirportMatches(entity interface{}, airport string) (city bool, ok bool) {
	code := at(entity, 0)
	kind := at(entity, 1)
	if code == airport && kind == float64(0) {
		return false, true
	}
	if airport == "BER" && code == "/m/0156q" && kind == float64(4) {
		return true, true
	}
	return false, false
}

func parsePage(body []byte, origin, destination, date string) ([]Flight, error) {
	context, err := pageData(body, "ds:0")
	if err != nil {
		return nil, err
	}
	filters := at(context, 1, 1)
	// Verify Google's interpretation of the free-text search before showing
	// fares. This rejects default dates, round trips and airport substitutions.
	segments, ok := at(filters, 13).([]interface{})
	if !ok || len(segments) != 1 || at(filters, 2) != float64(2) ||
		at(filters, 5) != float64(1) || at(filters, 6, 0) != float64(1) ||
		at(segments, 0, 6) != date {
		return nil, fmt.Errorf("search page does not match requested one-way economy itinerary")
	}
	originCity, originOK := pageAirportMatches(at(segments, 0, 0, 0, 0), origin)
	destinationCity, destinationOK := pageAirportMatches(at(segments, 0, 1, 0, 0), destination)
	if !originOK || !destinationOK {
		return nil, fmt.Errorf("search page does not match requested one-way economy itinerary")
	}
	data, err := pageData(body, "ds:1")
	if err != nil {
		return nil, err
	}
	flights, err := parseResults(data)
	if err != nil || (!originCity && !destinationCity) {
		return flights, err
	}
	// A city query may include nearby airports. Never show one as a BER fare.
	var matched []Flight
	for _, flight := range flights {
		if flight.Legs[0].DepAirport == origin && flight.Legs[len(flight.Legs)-1].ArrAirport == destination {
			matched = append(matched, flight)
		}
	}
	if len(flights) > 0 && len(matched) == 0 {
		return nil, fmt.Errorf("search page city results do not match requested airports")
	}
	return matched, nil
}

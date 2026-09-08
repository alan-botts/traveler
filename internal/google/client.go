package google

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

const (
	flightsURL = "https://www.google.com/_/FlightsFrontendUi/data/travel.frontend.flights.FlightsFrontendService/GetShoppingResults?hl=en&gl=US&curr=USD"
	maxRPS     = 10
)

// Client wraps a TLS-fingerprinted HTTP client for Google Flights requests.
type Client struct {
	httpClient tls_client.HttpClient
	mu         sync.Mutex
	lastReq    time.Time
}

// NewClient creates a Client with Chrome TLS fingerprint.
func NewClient() (*Client, error) {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_131),
		tls_client.WithCookieJar(jar),
		tls_client.WithNotFollowRedirects(),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS client: %w", err)
	}

	return &Client{
		httpClient: client,
	}, nil
}

// SearchFlights searches for one-way flights and returns parsed results.
func (c *Client) SearchFlights(origin, destination, date string) ([]Flight, error) {
	flights, rpcErr := c.searchRPC(origin, destination, date)
	if rpcErr == nil {
		return flights, nil
	}
	flights, pageErr := c.searchPage(origin, destination, date)
	if pageErr != nil {
		return nil, fmt.Errorf("RPC: %v; search page: %w", rpcErr, pageErr)
	}
	return flights, nil
}

func (c *Client) searchRPC(origin, destination, date string) ([]Flight, error) {
	c.rateLimit()

	body, err := BuildRequestBody(origin, destination, date)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req, err := http.NewRequest("POST", flightsURL, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://www.google.com")
	req.Header.Set("Referer", "https://www.google.com/travel/flights")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return ParseResponse(respBody)
}

// rateLimit ensures we don't exceed maxRPS requests per second.
func (c *Client) rateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()

	minInterval := time.Second / time.Duration(maxRPS)
	elapsed := time.Since(c.lastReq)
	if elapsed < minInterval {
		time.Sleep(minInterval - elapsed)
	}
	c.lastReq = time.Now()
}

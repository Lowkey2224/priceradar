package scraper

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// roundTripFunc lets tests stub an http.Client's transport, so no request leaves the process.
type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// clientReturning answers every request with status and body.
func clientReturning(status int, body string) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}
}

// clientFailing simulates a transport error, e.g. a dropped connection.
func clientFailing(err error) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, err
	})}
}

// clientRejectingRequests fails the test as soon as any request is sent.
func clientRejectingRequests(t *testing.T) *http.Client {
	t.Helper()
	return &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Errorf("unexpected HTTP request to %s", req.URL)
		return nil, errors.New("this client must not be used")
	})}
}

type requestLog struct {
	urls []string
}

// A bare injected client would replace CheckRedirect and test nothing.
func clientKeepingRedirectPolicy(log *requestLog, respond func(req *http.Request) *http.Response) *http.Client {
	client := AmazonScraper{}.httpClient()
	client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		log.urls = append(log.urls, req.URL.String())
		return respond(req), nil
	})

	return client
}

func redirectTo(req *http.Request, location string) *http.Response {
	header := make(http.Header)
	header.Set("Location", location)

	return &http.Response{
		StatusCode: http.StatusMovedPermanently,
		Status:     "301 Moved Permanently",
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     header,
		Request:    req,
	}
}

func responseOK(req *http.Request, body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    req,
	}
}

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("fixture %q is not readable: %v", name, err)
	}
	return string(content)
}

func TestSupports(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid Amazon.de URL",
			url:      "https://www.amazon.de/dp/B08N5WRWNW/",
			expected: true,
		},
		{
			name:     "uppercase scheme and host",
			url:      "HTTPS://WWW.AMAZON.DE/dp/B08N5WRWNW/",
			expected: true,
		},
		{
			name:     "explicit default port",
			url:      "https://www.amazon.de:443/dp/B08N5WRWNW/",
			expected: true,
		},
		{
			name:     "Amazon.com URL",
			url:      "https://www.amazon.com/dp/B08N5WRWNW/",
			expected: false,
		},
		{
			name:     "foreign host with Amazon.de URL in query",
			url:      "https://attacker.example/?next=https://www.amazon.de/dp/B08N5WRWNW",
			expected: false,
		},
		{
			name:     "unsupported URL (Google)",
			url:      "https://www.google.com",
			expected: false,
		},
		{
			name:     "empty URL",
			url:      "",
			expected: false,
		},
	}

	scraper := AmazonScraper{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scraper.supports(tt.url)
			if got != tt.expected {
				t.Errorf("supports(%q) = %v; want %v", tt.url, got, tt.expected)
			}
		})
	}
}

// TestParsePrices covers the selector and sanitize logic - the part that
// breaks as soon as Amazon changes its markup. No HTTP involved.
func TestParsePrices(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		expected    float64
		expectedErr error
	}{
		{
			name:     "thousands separator and decimal comma",
			html:     `<span id="apex-core-price-identifier"><span class="a-price-whole">1.117<span class="a-price-decimal">,</span></span><span class="a-price-fraction">45</span></span>`,
			expected: 1117.45,
		},
		{
			name:     "price without thousands separator",
			html:     `<span id="apex-core-price-identifier"><span class="a-price-whole">576<span class="a-price-decimal">,</span></span><span class="a-price-fraction">11</span></span>`,
			expected: 576.11,
		},
		{
			name:        "no price in markup",
			html:        `<div id="availability">Derzeit nicht verfügbar</div>`,
			expected:    0,
			expectedErr: ErrPriceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePrices(strings.NewReader(tt.html))
			if tt.expectedErr != nil && errors.Is(err, tt.expectedErr) == false {
				t.Fatalf("parsePrices() unexpected error: %v; want %v", err, tt.expectedErr)
			}

			if got != tt.expected {
				t.Errorf("parsePrices() = %f; want %f",
					got, tt.expected)
			}
		})
	}
}

func TestHTTPClient(t *testing.T) {
	t.Run("without injection the production client applies", func(t *testing.T) {
		got := AmazonScraper{}.httpClient()
		if got == nil {
			t.Fatal("httpClient() = nil; want client with timeout")
		}
		if got.Timeout != 10*time.Second {
			t.Errorf("Timeout = %v; want %v", got.Timeout, 10*time.Second)
		}
	})

	t.Run("injected client wins", func(t *testing.T) {
		injected := clientReturning(http.StatusOK, "")
		if got := (AmazonScraper{client: injected}).httpClient(); got != injected {
			t.Error("httpClient() does not return the injected client")
		}
	})
}

func TestScrape(t *testing.T) {
	fixture := loadFixture(t, "amazon-dp-B0FMS9XQF7.html")

	tests := []struct {
		name     string
		url      string
		status   int
		body     string
		expected float64
		wantErr  bool
	}{
		{
			name:     "valid Amazon.de URL",
			url:      "https://www.amazon.de/dp/B0FMS9XQF7",
			status:   http.StatusOK,
			body:     fixture,
			expected: 576.11,
		},
		{
			name:    "Amazon responds with 503",
			url:     "https://www.amazon.de/dp/B0FMS9XQF7",
			status:  http.StatusServiceUnavailable,
			body:    "",
			wantErr: true,
		},
		{
			name:    "page without price",
			url:     "https://www.amazon.de/dp/B0FMS9XQF7",
			status:  http.StatusOK,
			body:    `<div id="availability">Derzeit nicht verfügbar</div>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraper := AmazonScraper{client: clientReturning(tt.status, tt.body)}

			got, err := scraper.scrape(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("scrape(%q) = %v, nil; want error", tt.url, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("scrape(%q) unexpected error: %v", tt.url, err)
			}
			if got != tt.expected {
				t.Errorf("scrape(%q) = %v; want %v", tt.url, got, tt.expected)
			}
		})
	}

	t.Run("unsupported URL triggers no request", func(t *testing.T) {
		scraper := AmazonScraper{client: clientRejectingRequests(t)}

		got, err := scraper.scrape("https://www.amazon.com/dp/B0FMS9XQF7")
		if err == nil {
			t.Fatalf("scrape() = %v, nil; want error", got)
		}
	})
}

func TestFetchPrice(t *testing.T) {
	fixture := loadFixture(t, "amazon-dp-B0FMS9XQF7.html")

	t.Run("price from the response", func(t *testing.T) {
		scraper := AmazonScraper{client: clientReturning(http.StatusOK, fixture)}

		got, err := scraper.fetchPrices("https://www.amazon.de/dp/B0FMS9XQF7")
		if err != nil {
			t.Fatalf("fetchPrices() unexpected error: %v", err)
		}
		if got != 576.11 {
			t.Errorf("fetchPrices() = %f; want %f", got, 576.11)
		}
	})

	t.Run("status other than 200 is reported", func(t *testing.T) {
		scraper := AmazonScraper{client: clientReturning(http.StatusServiceUnavailable, "")}

		_, err := scraper.fetchPrices("https://www.amazon.de/dp/B0FMS9XQF7")
		if err == nil {
			t.Fatal("fetchPrices() = nil; want error because of status 503")
		}
		if !strings.Contains(err.Error(), "503") {
			t.Errorf("error does not name the status: %v", err)
		}
	})

	t.Run("transport error is passed through", func(t *testing.T) {
		wantErr := errors.New("connection dropped")
		scraper := AmazonScraper{client: clientFailing(wantErr)}

		_, err := scraper.fetchPrices("https://www.amazon.de/dp/B0FMS9XQF7")
		if !errors.Is(err, wantErr) {
			t.Errorf("fetchPrices() = %v; want %v", err, wantErr)
		}
	})

	t.Run("browser headers are set", func(t *testing.T) {
		var seen http.Header
		scraper := AmazonScraper{client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				seen = req.Header.Clone()
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(strings.NewReader(fixture)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}),
		}}

		if _, err := scraper.fetchPrices("https://www.amazon.de/dp/B0FMS9XQF7"); err != nil {
			t.Fatalf("fetchPrices() unexpected error: %v", err)
		}
		if !strings.Contains(seen.Get("User-Agent"), "Mozilla/5.0") {
			t.Errorf("User-Agent = %q; want browser identifier", seen.Get("User-Agent"))
		}
		if !strings.HasPrefix(seen.Get("Accept-Language"), "de-DE") {
			t.Errorf("Accept-Language = %q; want de-DE first", seen.Get("Accept-Language"))
		}
	})
}

func TestRedirectPolicy(t *testing.T) {
	const start = "https://www.amazon.de/dp/B0FMS9XQF7"
	fixture := loadFixture(t, "amazon-dp-B0FMS9XQF7.html")

	respondWithRedirect := func(location string) func(req *http.Request) *http.Response {
		return func(req *http.Request) *http.Response {
			if req.URL.String() == start {
				return redirectTo(req, location)
			}

			return responseOK(req, fixture)
		}
	}

	offSite := []struct {
		name     string
		location string
	}{
		{name: "foreign host", location: "https://attacker.example/dp/B0FMS9XQF7"},
		{name: "http downgrade", location: "http://www.amazon.de/dp/B0FMS9XQF7"},
		{name: "lookalike host", location: "https://www.amazon.de.attacker.example/dp/B0FMS9XQF7"},
	}

	for _, tt := range offSite {
		t.Run(tt.name+" redirect is not followed", func(t *testing.T) {
			var log requestLog
			scraper := AmazonScraper{client: clientKeepingRedirectPolicy(&log, respondWithRedirect(tt.location))}

			if _, err := scraper.fetchPrices(start); !errors.Is(err, ErrRedirectNotAllowed) {
				t.Fatalf("fetchPrices() error = %v; want %v", err, ErrRedirectNotAllowed)
			}
			if len(log.urls) != 1 || log.urls[0] != start {
				t.Errorf("transport saw %v; want only %q", log.urls, start)
			}
		})
	}

	t.Run("redirect within amazon.de is followed", func(t *testing.T) {
		const canonical = "https://www.amazon.de/Some-Product-Name/dp/B0FMS9XQF7"
		var log requestLog
		scraper := AmazonScraper{client: clientKeepingRedirectPolicy(&log, respondWithRedirect(canonical))}

		got, err := scraper.fetchPrices(start)
		if err != nil {
			t.Fatalf("fetchPrices() unexpected error: %v", err)
		}
		if got != 576.11 {
			t.Errorf("fetchPrices() = %f; want %f", got, 576.11)
		}
		if len(log.urls) != 2 || log.urls[1] != canonical {
			t.Errorf("transport saw %v; want %q then %q", log.urls, start, canonical)
		}
	})
}

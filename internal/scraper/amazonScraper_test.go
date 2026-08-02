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

// roundTripFunc erlaubt es, einen http.Client ohne Netzwerkzugriff zu bestücken.
type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// clientReturning antwortet auf jede Anfrage mit status und body.
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

// clientFailing simuliert einen Transportfehler, etwa eine abgebrochene Verbindung.
func clientFailing(err error) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, err
	})}
}

// clientRejectingRequests lässt den Test fehlschlagen, sobald überhaupt ein
// Request abgesetzt wird.
func clientRejectingRequests(t *testing.T) *http.Client {
	t.Helper()
	return &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Errorf("unerwarteter HTTP-Request an %s", req.URL)
		return nil, errors.New("dieser Client darf nicht verwendet werden")
	})}
}

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("Fixture %q ist nicht lesbar: %v", name, err)
	}
	return string(content)
}

func TestSupports(t *testing.T) {
	// 1. Die Testtabelle (DataProvider) als Slice von Structs
	tests := []struct {
		name     string // Name des Testfalls
		url      string // Input
		expected bool   // Erwartetes Ergebnis
	}{
		{
			name:     "Gültige Amazon.de URL",
			url:      "https://www.amazon.de/dp/B08N5WRWNW/",
			expected: true,
		},
		{
			name:     "Gültige Amazon.com URL",
			url:      "https://www.amazon.com/dp/B08N5WRWNW/",
			expected: false,
		},
		{
			name:     "Gültige Amazon.com URL",
			url:      "https://attacker.example/?next=https://www.amazon.de/dp/B08N5WRWNW",
			expected: false,
		},
		{
			name:     "Ungültige URL (Google)",
			url:      "https://www.google.com",
			expected: false,
		},
		{
			name:     "Leere URL",
			url:      "",
			expected: false,
		},
	}

	scraper := AmazonScraper{}

	// 2. Iteration über die Testfälle
	for _, tt := range tests {
		// t.Run führt jeden Fall als isolierten Unter-Test aus
		t.Run(tt.name, func(t *testing.T) {
			got := scraper.supports(tt.url)
			if got != tt.expected {
				t.Errorf("supports(%q) = %v; want %v", tt.url, got, tt.expected)
			}
		})
	}
}

// TestParsePrices deckt die Selektor- und Sanitize-Logik ab - den Teil, der
// bricht, sobald Amazon sein Markup ändert. Kein HTTP im Spiel.
func TestParsePrices(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		expected    float64
		expectedErr error
	}{
		{
			name:     "Tausenderpunkt und Dezimalkomma",
			html:     `<span class="a-price-whole">1.117<span class="a-price-decimal">,</span></span><span class="a-price-fraction">45</span>`,
			expected: 1117.45,
		},
		{
			name:     "Preis ohne Tausendertrenner",
			html:     `<span class="a-price-whole">576<span class="a-price-decimal">,</span></span><span class="a-price-fraction">11</span>`,
			expected: 576.11,
		},
		{
			name: "Erster Preisblock gewinnt",
			html: `<span class="a-price-whole">576<span class="a-price-decimal">,</span></span><span class="a-price-fraction">11</span>` +
				`<span class="a-price-whole">999<span class="a-price-decimal">,</span></span><span class="a-price-fraction">99</span>`,
			expected: 576.11,
		},
		{
			name:        "Kein Preis im Markup",
			html:        `<div id="availability">Derzeit nicht verfügbar</div>`,
			expected:    0,
			expectedErr: ErrPriceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePrices(strings.NewReader(tt.html))
			if tt.expectedErr != nil && errors.Is(err, tt.expectedErr) == false {
				t.Fatalf("parsePrices() unerwarteter Fehler: %v; want %v", err, tt.expectedErr)
			}

			if got != tt.expected {
				t.Errorf("parsePrices() = %f; want %f",
					got, tt.expected)
			}
		})
	}
}

func TestHTTPClient(t *testing.T) {
	t.Run("ohne Injektion greift der Produktionsclient", func(t *testing.T) {
		got := AmazonScraper{}.httpClient()
		if got == nil {
			t.Fatal("httpClient() = nil; want Client mit Timeout")
		}
		if got.Timeout != 10*time.Second {
			t.Errorf("Timeout = %v; want %v", got.Timeout, 10*time.Second)
		}
	})

	t.Run("injizierter Client gewinnt", func(t *testing.T) {
		injected := clientReturning(http.StatusOK, "")
		if got := (AmazonScraper{client: injected}).httpClient(); got != injected {
			t.Error("httpClient() liefert nicht den injizierten Client")
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
			name:     "Gültige Amazon.de URL",
			url:      "https://www.amazon.de/dp/B0FMS9XQF7",
			status:   http.StatusOK,
			body:     fixture,
			expected: 576.11,
		},
		{
			name:    "Amazon antwortet mit 503",
			url:     "https://www.amazon.de/dp/B0FMS9XQF7",
			status:  http.StatusServiceUnavailable,
			body:    "",
			wantErr: true,
		},
		{
			name:    "Seite ohne Preis",
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
					t.Fatalf("scrape(%q) = %v, nil; want Fehler", tt.url, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("scrape(%q) unerwarteter Fehler: %v", tt.url, err)
			}
			if got != tt.expected {
				t.Errorf("scrape(%q) = %v; want %v", tt.url, got, tt.expected)
			}
		})
	}

	t.Run("Nicht unterstützte URL löst keinen Request aus", func(t *testing.T) {
		scraper := AmazonScraper{client: clientRejectingRequests(t)}

		got, err := scraper.scrape("https://www.amazon.com/dp/B0FMS9XQF7")
		if err == nil {
			t.Fatalf("scrape() = %v, nil; want Fehler", got)
		}
	})
}

func TestFetchPrice(t *testing.T) {
	fixture := loadFixture(t, "amazon-dp-B0FMS9XQF7.html")

	t.Run("Preis aus der Antwort", func(t *testing.T) {
		scraper := AmazonScraper{client: clientReturning(http.StatusOK, fixture)}

		got, err := scraper.fetchPrices("https://www.amazon.de/dp/B0FMS9XQF7")
		if err != nil {
			t.Fatalf("fetchPrices() unerwarteter Fehler: %v", err)
		}
		if got != 576.11 {
			t.Errorf("fetchPrices() = %f; want %f", got, 576.11)
		}
	})

	t.Run("Status ungleich 200 wird gemeldet", func(t *testing.T) {
		scraper := AmazonScraper{client: clientReturning(http.StatusServiceUnavailable, "")}

		_, err := scraper.fetchPrices("https://www.amazon.de/dp/B0FMS9XQF7")
		if err == nil {
			t.Fatal("fetchPrices() = nil; want Fehler wegen Status 503")
		}
		if !strings.Contains(err.Error(), "503") {
			t.Errorf("Fehler nennt den Status nicht: %v", err)
		}
	})

	t.Run("Transportfehler wird durchgereicht", func(t *testing.T) {
		wantErr := errors.New("Verbindung abgebrochen")
		scraper := AmazonScraper{client: clientFailing(wantErr)}

		_, err := scraper.fetchPrices("https://www.amazon.de/dp/B0FMS9XQF7")
		if !errors.Is(err, wantErr) {
			t.Errorf("fetchPrices() = %v; want %v", err, wantErr)
		}
	})

	t.Run("Browser-Header werden gesetzt", func(t *testing.T) {
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
			t.Fatalf("fetchPrices() unerwarteter Fehler: %v", err)
		}
		if !strings.Contains(seen.Get("User-Agent"), "Mozilla/5.0") {
			t.Errorf("User-Agent = %q; want Browser-Kennung", seen.Get("User-Agent"))
		}
		if !strings.HasPrefix(seen.Get("Accept-Language"), "de-DE") {
			t.Errorf("Accept-Language = %q; want de-DE zuerst", seen.Get("Accept-Language"))
		}
	})
}

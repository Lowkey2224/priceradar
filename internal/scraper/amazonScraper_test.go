package scraper

import "testing"

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

func TestScrape(t *testing.T) {
	tests := []struct {
		name     string  // Name des Testfalls
		url      string  // Input
		expected float64 // Erwartetes Ergebnis
	}{
		{
			name:     "Gültige Amazon.de URL",
			url:      "https://www.amazon.de/dp/B0GFDPHRXS",
			expected: 1117.45,
		},
		{
			name:     "Gültige Amazon.de URL",
			url:      "https://www.amazon.de/dp/B0FMS9XQF7",
			expected: 576.11,
		},
	}

	scraper := AmazonScraper{}

	// 2. Iteration über die Testfälle
	for _, tt := range tests {
		// t.Run führt jeden Fall als isolierten Unter-Test aus
		t.Run(tt.name, func(t *testing.T) {
			got, err := scraper.scrape(tt.url)
			if got != tt.expected || err != nil {
				t.Errorf("scrape(%q) = %v, %v; want %v", tt.url, got, err, tt.expected)
			}
		})
	}
}

func TestFetchPrice(t *testing.T) {
	tests := []struct {
		name             string // Name des Testfalls
		url              string // Input
		expectedIntegers string // Erwartetes Ergebnis
		expectedDecimals string
	}{
		{
			name:             "Gültige Amazon.de URL",
			url:              "https://www.amazon.de/ASUS-Prime-Radeon-Gaming-Grafikkarte/dp/B0FMS9XQF7?dib=eyJ2IjoiMSJ9.Ahed929b2TPNEEcFrq8g0JnLrW8KIbUTFglx1MXUwGpRBpDeQLEjc5h9DJMTqAlVnLSbY5J1gmemjwnEFdQCSwKsvn50Z6WaDoQC1Y1UGPKm6UzbiyfaQ1oQGxzs5eAtklRzAse1QPuSfCz2fKm9_wgwMKJ32RkaR-XNhXBaFTcYB-mfg5Qd1vf9PjMIfXO0O1_YywP1fchDafr7bNUdUuRP5A1-yEQlIcvN9X6H8PY.-V_VJZyv-MJEFffZSMtQd4YmpJ5ai_owM9wb15bNJ0g&dib_tag=se&keywords=rx%2B9070&qid=1785399398&sr=8-3&th=1",
			expectedIntegers: "576",
			expectedDecimals: "11",
		},
		{
			name:             "Gültige Amazon.de URL",
			url:              "https://www.amazon.de/ASUS-Prime-Radeon-Gaming-Grafikkarte/dp/B0FMS9XQF7",
			expectedIntegers: "576",
			expectedDecimals: "11",
		},
		{
			name:             "Gültige Amazon.de URL",
			url:              "https://www.amazon.de/dp/B0FMS9XQF7",
			expectedIntegers: "576",
			expectedDecimals: "11",
		},
	}

	scraper := AmazonScraper{}

	// 2. Iteration über die Testfälle
	for _, tt := range tests {
		// t.Run führt jeden Fall als isolierten Unter-Test aus
		t.Run(tt.name, func(t *testing.T) {
			gotInt, gotDec, err := scraper.fetchPrices(tt.url)
			if gotInt != tt.expectedIntegers || gotDec != tt.expectedDecimals || err != nil {
				t.Errorf("scrape(%q) = %v, %v, %v; want %v, %v", tt.url, gotInt, gotDec, err, tt.expectedIntegers, tt.expectedDecimals)
			}
		})
	}
}

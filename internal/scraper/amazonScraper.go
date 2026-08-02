package scraper

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type AmazonScraper struct {
	// client ist im Normalbetrieb nil; Tests setzen hier einen Client mit
	// eigenem Transport, um ohne Netzwerkzugriff zu antworten.
	client *http.Client
}

var amazonProductPath = regexp.MustCompile(`^/(?:[^/]+/)?dp/[A-Za-z0-9]{10}(?:/|$)`)
var ErrPriceNotFound = errors.New("price not found")

func (s AmazonScraper) scrape(url string) (float64, error) {
	if !s.supports(url) {
		return 0, errors.New("URL is not supported")
	}

	return s.fetchPrices(url)
}

func (s AmazonScraper) fetchPrices(url string) (float64, error) {
	res, err := s.fetchHTML(url)

	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := res.Body.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	if res.StatusCode != 200 {
		return 0, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	return parsePrices(res.Body)
}

func parsePrices(r io.Reader) (float64, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return 0, err
	}
	priceElem := doc.Find(".a-price").First()
	whole := priceElem.Find(".a-price-whole").First()

	decimals := priceElem.Find(".a-price-fraction").First().Text()
	sanitized := strings.ReplaceAll(whole.Text(), ",", "")
	sanitized = strings.ReplaceAll(sanitized, ".", "")
	if sanitized == "" || decimals == "" {
		return 0, fmt.Errorf("%w: integer: %v, decimal: %v", ErrPriceNotFound, sanitized, decimals)
	}
	return strconv.ParseFloat(fmt.Sprintf("%s.%s", sanitized, decimals), 64)
}

func (s AmazonScraper) supports(url string) bool {
	parsed, err := urlpkg.Parse(url)
	return err == nil &&
		parsed.Scheme == "https" &&
		strings.EqualFold(parsed.Hostname(), "www.amazon.de") &&
		amazonProductPath.MatchString(parsed.EscapedPath())
}

func (s AmazonScraper) scraperName() string {
	return "Amazon.de"
}

// httpClient liefert den injizierten Client oder den Produktionsclient
// (mit Timeout ist immer Best Practice!).
func (s AmazonScraper) httpClient() *http.Client {
	if s.client != nil {
		return s.client
	}

	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

func (s AmazonScraper) fetchHTML(url string) (*http.Response, error) {
	// 2. Request-Objekt anlegen
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 3. Browser-Header setzen (User-Agent vortäuschen)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Optional: Weitere typische Browser-Header mitsenden
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	// 4. Request absenden
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

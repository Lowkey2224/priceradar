package scraper

import (
	"errors"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type AmazonScraper struct {
}

const baseUrl = "https://www.amazon.de/"

var amazonProductPath = regexp.MustCompile(`^/(?:[^/]+/)?dp/[A-Za-z0-9]{10}(?:/|$)`)

func (s AmazonScraper) scrape(url string) (float64, error) {
	if s.supports(url) == false {
		return 0, errors.New("URL is not supported")
	}

	intPrice, decimals, err := s.fetchPrices(url)
	if err != nil {
		return 0.0, err
	}
	price, err := strconv.ParseFloat(fmt.Sprintf("%s.%s", intPrice, decimals), 64)
	return price, err
}

func (s AmazonScraper) fetchPrices(url string) (integers, decimals string, err error) {
	res, err := fetchHTML(url)

	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	//TODO Fetch content of html element with some xpath selector

	if res.StatusCode != 200 {
		return "", "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", "", err
	}

	whole := doc.Find(".a-price-whole").First()
	// Find the review items
	integers = whole.Text()
	decimals = doc.Find(".a-price-fraction").First().Text()
	sanitized := strings.ReplaceAll(integers, ",", "")
	sanitized = strings.ReplaceAll(sanitized, ".", "")
	return sanitized, decimals, nil
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

func fetchHTML(url string) (*http.Response, error) {
	// 1. Einen HTTP-Client erstellen (mit Timeout ist immer Best Practice!)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

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
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

package scraper

type Scraper struct {
	url string
}

type ScraperInterface interface {
	scrape(url string) (float64, error)
	supports(url string) bool
	scraperName() string
}

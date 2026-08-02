package scraper

type Scraper struct {
}

type ScraperInterface interface {
	scrape(url string) (float64, error)
	supports(url string) bool
	scraperName() string
}

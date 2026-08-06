package scraper

type ScraperManager struct {
	scrapers []ScraperInterface
}

func Create() ScraperManager {
	return ScraperManager{
		scrapers: []ScraperInterface{
			AmazonScraper{},
		},
	}
}

func (man ScraperManager) GetPrice(url string) (map[string]float64, error) {
	priceMap := make(map[string]float64)
	for _, scraper := range man.scrapers {
		if scraper.supports(url) {
			price, err := scraper.scrape(url)
			if err != nil {
				return priceMap, err
			}
			priceMap[scraper.scraperName()] = price
		}
	}
	return priceMap, nil
}

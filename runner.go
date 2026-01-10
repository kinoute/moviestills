package main

import (
	"moviestills/config"
	"moviestills/debug"
	"moviestills/scraper"
	"path/filepath"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/pterm/pterm"
)

func runSequential(websitesToScrape []string, options *config.Options, aggStats *scraper.AggregatedStats) {
	for _, website := range websitesToScrape {
		pterm.DefaultSection.Println("Scraping", website)
		stats := runScraper(website, options)
		aggStats.Add(stats)
	}
}

func runConcurrent(websitesToScrape []string, options *config.Options, aggStats *scraper.AggregatedStats) {
	pterm.Info.Println("Running", pterm.White(len(websitesToScrape)), "scrapers concurrently...")

	var wg sync.WaitGroup
	for _, website := range websitesToScrape {
		wg.Add(1)
		go func(site string) {
			defer wg.Done()
			stats := runScraper(site, options)
			aggStats.Add(stats)
		}(website)
	}
	wg.Wait()
}

func runScraper(website string, options *config.Options) *scraper.Stats {
	// Create and configure scraper for this website
	c := colly.NewCollector(
		colly.CacheDir(filepath.Join(options.CacheDir, website)),
	)

	configureScraper(c, options)

	// Initialize stats tracking
	stats := &scraper.Stats{Website: website}

	// Run the scraper
	siteFunc := sites[website]
	siteFunc(c, options, stats)

	pterm.Success.Println("Finished scraping", pterm.White(website))

	return stats
}

func configureScraper(c *colly.Collector, options *config.Options) {
	// Set up a proxy
	if options.Proxy != "" {
		if err := c.SetProxy(options.Proxy); err != nil {
			pterm.Error.Println("Could not set proxy", pterm.White(options.Proxy), pterm.Red(err))
		}
	}

	// Set request timeout
	c.SetRequestTimeout(options.TimeOut)

	// Enable asynchronous jobs if asked
	if options.Async {
		c.Async = true
	}

	// Enable Debugging level if asked through the CLI
	if options.Debug {
		pterm.EnableDebugMessages()
		c.SetDebugger(&debug.PTermDebugger{})
	}

	// Use random user agent and referer to avoid getting banned
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	// Limit parallelism and add random delay to avoid getting IP banned
	if err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: options.Parallel,
		RandomDelay: options.RandomDelay,
	}); err != nil {
		pterm.Warning.Println("Can't change scraper limit options:", pterm.Red(err))
	}
}

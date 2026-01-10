package main

import (
	"moviestills/config"
	"moviestills/scraper"
	"moviestills/websites"
	"os"
	"sort"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/gocolly/colly/v2"
	"github.com/pterm/pterm"
)

// ScraperFunc is the function signature for all website scrapers
type ScraperFunc func(*colly.Collector, *config.Options, *scraper.Stats)

// Available scrapers
var sites = map[string]ScraperFunc{
	"blubeaver":        websites.BluBeaverScraper,
	"blusscreens":      websites.BlusScraper,
	"dvdbeaver":        websites.DVDBeaverScraper,
	"evanerichards":    websites.EvanERichardsScraper,
	"film-grab":        websites.FilmGrabScraper,
	"highdefdiscnews":  websites.HighDefDiscNewsScraper,
	"movie-screencaps": websites.ScreenCapsScraper,
	"screenmusings":    websites.ScreenMusingsScraper,
	"stillsfrmfilms":   websites.StillsFrmFilmsScraper,
}

func main() {
	// Start by cleaning the Terminal Screen
	clearScreen()

	// Stop program when user is pressing CTRL+C
	handleShutdown()

	// Handle arguments passed through the CLI or environment variables
	var options config.Options
	arg.MustParse(&options)

	// Interface of the app
	pterm.DefaultHeader.Println("Movie Stills", config.VERSION)

	// Adjust logging styles
	setupLogging(&options)

	// Display available scrapers implemented
	if options.ListScrapers {
		listAvailableScrapers()
		return
	}

	// Determine which websites to scrape
	websitesToScrape := determineWebsites(&options)
	if len(websitesToScrape) == 0 {
		pterm.Error.Println("No website specified. Use --website or --all flag.")
		listAvailableScrapers()
		os.Exit(1)
	}

	// Validate all specified websites exist
	for _, website := range websitesToScrape {
		if _, exists := sites[website]; !exists {
			pterm.Error.Println("We don't have a scraper for:", pterm.White(website))
			listAvailableScrapers()
			os.Exit(1)
		}
	}

	pterm.DefaultSection.Println("Configuration")
	printConfiguration(&options, websitesToScrape)

	// Create the necessary directories (cache and data)
	setupDirectories(&options)

	// Run scrapers
	aggStats := scraper.NewAggregatedStats()

	if len(websitesToScrape) == 1 || options.Sequential {
		runSequential(websitesToScrape, &options, aggStats)
	} else {
		runConcurrent(websitesToScrape, &options, aggStats)
	}

	// Print final summary
	pterm.Info.Println("Finished scraping", pterm.White(len(websitesToScrape)), "website(s)")
	scraper.PrintAggregatedSummary(aggStats)
}

func determineWebsites(options *config.Options) []string {
	if options.All {
		all := make([]string, 0, len(sites))
		for name := range sites {
			all = append(all, name)
		}
		sort.Strings(all)
		return all
	}

	// Normalize and deduplicate website list
	seen := make(map[string]bool)
	result := make([]string, 0, len(options.Website))
	for _, w := range options.Website {
		w = strings.ToLower(strings.TrimSpace(w))
		if w != "" && !seen[w] {
			seen[w] = true
			result = append(result, w)
		}
	}
	return result
}

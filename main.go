package main

import (
	"log"
	"moviestills/config"
	"moviestills/utils"
	"moviestills/websites"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/gocolly/colly/v2/extensions"
)

func main() {

	// Implemented scrapers as today
	sites := map[string]func(**colly.Collector, *config.Options){
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

	// Handle arguments passed through the CLI or environment variables
	// Check config/config.go for the list of available options.
	var options config.Options
	arg.MustParse(&options)

	// Check presence of website argument
	if options.Website == "" {
		log.Fatalln("A website must be set through arguments.")
	}

	// Verify if we have a scrapper for the given website.
	// If we do, "site_func" will now contain a function listed in
	// the sites map that matches a module for this specific
	// website stored in the "websites" folder.
	site_func, scraper_exists := sites[strings.ToLower(options.Website)]
	if !scraper_exists {
		log.Println("We don't have a scraper for this website.")
		log.Println("List of available scrapers:")
		for site := range sites {
			log.Println("–", site)
		}
		log.Fatalln("See how you can add support for a new website: https://github.com/kinoute/moviestills#contribute")
	}

	// If we're here, it means we have a valid scraper for a valid website!

	// Create the "cache" directory.
	// This folder stores the scraped websites pages.
	// If we can't create it, stop right there.
	if _, err := utils.CreateFolder(options.CacheDir); err != nil {
		log.Fatalln("The cache directory", options.CacheDir, "can't be created:", err)
	}

	// Create the "data" directory.
	// This folder stores the movie snapshots.
	// If we can't create it, stop right there.
	if _, err := utils.CreateFolder(options.DataDir); err != nil {
		log.Fatalln("The data directory", options.DataDir, "can't be created:", err)
	}

	// Instantiate main scraper
	scraper := colly.NewCollector(
		colly.CacheDir(options.CacheDir),
	)

	// Enable asynchronous jobs if asked
	if options.Async {
		scraper.Async = true
	}

	// Enable Colly Debugging if asked through the CLI
	if options.Debug {
		scraper.SetDebugger(&debug.LogDebugger{})
	}

	// Use random user agent and referer to avoid getting banned
	extensions.RandomUserAgent(scraper)
	extensions.Referer(scraper)

	// Limit parallelism and add random delay to avoid getting IP banned
	if err := scraper.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: options.Parallel,
	}); err != nil {
		log.Println("Can't change scraper limit options:", err)
	}

	// Here we call the website module depending on the website provided
	// in the CLI by the user.
	// This will call a file/module/func made specifically to scrap this website.
	// All available scrapers are stored in the "websites" folder.
	site_func(&scraper, &options)

}

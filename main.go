package main

import (
	"moviestills/config"
	"moviestills/utils"
	"moviestills/websites"
	"os"
	"reflect"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/gocolly/colly/v2/extensions"
	log "github.com/pterm/pterm"
)

func main() {

	clearScreen()

	// Implemented scrapers as today
	sites := map[string]func(**colly.Collector, *config.Options){
		"blubeaver": websites.BluBeaverScraper,
		// "blusscreens":      websites.BlusScraper,
		// "dvdbeaver":        websites.DVDBeaverScraper,
		// "evanerichards":    websites.EvanERichardsScraper,
		// "film-grab":        websites.FilmGrabScraper,
		// "highdefdiscnews":  websites.HighDefDiscNewsScraper,
		// "movie-screencaps": websites.ScreenCapsScraper,
		// "screenmusings":    websites.ScreenMusingsScraper,
		// "stillsfrmfilms":   websites.StillsFrmFilmsScraper,
	}

	// Handle arguments passed through the CLI or environment variables
	// Check config/config.go for the list of available options.
	var options config.Options
	arg.MustParse(&options)

	log.DefaultHeader.Println("Movie Stills", config.VERSION)

	log.DefaultSection.Println("Configuration")
	printConfiguration(&options)

	// Check presence of website argument
	if options.Website == "" {
		log.Error.Println("A website must be set through arguments.")
		os.Exit(1)
	}

	// Verify if we have a scrapper for the given website.
	// If we do, "site_func" will now contain a function listed in
	// the sites map that matches a module for this specific
	// website stored in the "websites" folder.
	site_func, scraper_exists := sites[strings.ToLower(options.Website)]
	if !scraper_exists {
		log.Error.Println("We don't have a scraper for this website.")
		log.Info.Println("List of available scrapers:")
		for site := range sites {
			log.Info.Println("–", site)
		}
		log.Info.Println("See how you can add support for a new website:",
			log.White("https://github.com/kinoute/moviestills#contribute"))
		os.Exit(1)
	}

	// If we're here, it means we have a valid scraper for a valid website!

	// Create the "cache" directory.
	// This folder stores the scraped websites pages.
	// If we can't create it, stop right there.
	if _, err := utils.CreateFolder(options.CacheDir); err != nil {
		log.Error.Println("The cache directory", log.White(options.CacheDir), "can't be created:", log.Red(err))
		os.Exit(1)
	}

	// Create the "data" directory.
	// This folder stores the movie snapshots.
	// If we can't create it, stop right there.
	if _, err := utils.CreateFolder(options.DataDir); err != nil {
		log.Error.Println("The data directory", log.White(options.DataDir), "can't be created:", log.Red(err))
		os.Exit(1)
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
		log.Warning.Println("Can't change scraper limit options:", log.Red(err))
	}

	log.DefaultSection.Println("Scraping")

	// Here we call the website module depending on the website provided
	// in the CLI by the user.
	// This will call a file/module/func made specifically to scrap this website.
	// All available scrapers are stored in the "websites" folder.
	site_func(&scraper, &options)

}

// Clear Terminal Screen before startnig the app
func clearScreen() {
	print("\033[H\033[2J")
}

// Print configuration as a bullet list
func printConfiguration(options *config.Options) {

	// Get fields and its values from the config struct
	values := reflect.ValueOf((*options))
	fields := values.Type()

	// Create bullet lists with configuration
	configuration := []log.BulletListItem{}

	for i := 0; i < values.NumField(); i++ {
		configuration = append(configuration,
			log.BulletListItem{
				Level:       0,
				Text:        log.Yellow(fields.Field(i).Name) + ": " + log.Blue((values.Field(i).Interface())),
				TextStyle:   log.NewStyle(log.FgBlue),
				BulletStyle: log.NewStyle(log.FgRed),
			},
		)
	}

	// Print the configuration as a bullet list
	err := log.DefaultBulletList.WithItems(configuration).Render()
	if err != nil {
		log.Error.Println("Could not print configuration", log.Red(err))
	}
}

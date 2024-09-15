package main

import (
	"moviestills/config"
	"moviestills/debug"
	"moviestills/utils"
	"moviestills/websites"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"

	"github.com/alexflint/go-arg"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	log "github.com/pterm/pterm"
)

func main() {

	// Start by cleaning the Terminal Screen
	clearScreen()

	// Stop program when user is pressing CTRL+C
	handleShutdown()

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

	// Interface of the app
	log.DefaultHeader.Println("Movie Stills", config.VERSION)

	// Adjust logging styles
	setupLogging(&options)

	// Display available scrapers implemented
	if options.ListScrapers {
		listAvailableScrapers(sites)
		return
	}

	// Check presence of website argument
	if options.Website == "" {
		log.Error.Println("A website must be set through arguments.")
		listAvailableScrapers(sites)
		os.Exit(1)
	}

	website := strings.ToLower(options.Website)

	// Verify if we have a scrapper for the given website.
	// If we do, "site_func" will now contain a function listed in
	// the sites map that matches a module for this specific
	// website stored in the "websites" folder.
	if _, exists := sites[website]; !exists {
		log.Error.Println("We don't have a scraper for:", log.White(website))
		listAvailableScrapers(sites)
		os.Exit(1)
	}

	log.DefaultSection.Println("Configuration")
	printConfiguration(&options)

	// Create the necessary directories (cache and data)
	setupDirectories(&options)

	// Create and configure scraper for each website
	scraper := colly.NewCollector(
		colly.CacheDir(filepath.Join(options.CacheDir, website)),
	)

	// Set up scraper settings
	configureScraper(&scraper, &options)

	// Run the scraper for the current website
	siteFunc := sites[website]
	siteFunc(&scraper, &options)

	log.Info.Println("Finished scraping", log.White(website))

}

func clearScreen() {
	print("\033[H\033[2J")
}

func handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		log.Info.Println("Shutting down...")
		os.Exit(130)
	}()
}

func setupLogging(options *config.Options) {
	// Adjust the logging prefix
	log.Info = *log.Info.WithPrefix(log.Prefix{Text: " INFOS ", Style: log.Info.Prefix.Style})

	// Disable styles and colors based on options
	if options.NoStyle {
		log.DisableStyling()
	}
	if options.NoColors {
		log.DisableColor()
	}
}

func setupDirectories(options *config.Options) {
	// Create the cache directory
	if _, err := utils.CreateFolder(options.CacheDir); err != nil {
		log.Error.Println("The cache directory", log.White(options.CacheDir), "can't be created:", log.Red(err))
		os.Exit(1)
	}

	// Create the data directory
	if _, err := utils.CreateFolder(options.DataDir); err != nil {
		log.Error.Println("The data directory", log.White(options.DataDir), "can't be created:", log.Red(err))
		os.Exit(1)
	}
}

func configureScraper(scraper **colly.Collector, options *config.Options) {
	// Set up a proxy
	if options.Proxy != "" {
		if err := (*scraper).SetProxy(options.Proxy); err != nil {
			log.Error.Println("Could not set proxy", log.White(options.Proxy), log.Red(err))
		}
	}

	// Set request timeout
	(*scraper).SetRequestTimeout(options.TimeOut)

	// Enable asynchronous jobs if asked
	if options.Async {
		(*scraper).Async = true
	}

	// Enable Debugging level if asked through the CLI
	if options.Debug {
		log.EnableDebugMessages()
		(*scraper).SetDebugger(&debug.PTermDebugger{})
	}

	// Use random user agent and referer to avoid getting banned
	extensions.RandomUserAgent((*scraper))
	extensions.Referer((*scraper))

	// Limit parallelism and add random delay to avoid getting IP banned
	if err := (*scraper).Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: options.Parallel,
		RandomDelay: 1 * options.RandomDelay,
	}); err != nil {
		log.Warning.Println("Can't change scraper limit options:", log.Red(err))
	}
}

// Print configuration as a bullet list. Most
// likely when the app starts.
func printConfiguration(options *config.Options) {

	// Get fields and its values from the config struct
	values := reflect.ValueOf(*options)
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
	if err := log.DefaultBulletList.WithItems(configuration).Render(); err != nil {
		log.Error.Println("Could not print configuration", log.Red(err))
	}
}

// Print list of available scrapers. We use it with
// the --list flag or when a user enters a website's name
// that is not implemented.
func listAvailableScrapers(sites map[string]func(**colly.Collector, *config.Options)) {

	log.DefaultSection.Println("Scrapers available")

	// Create bullet lists with available scrapers
	availableScrapers := []log.BulletListItem{}
	for site := range sites {
		availableScrapers = append(availableScrapers,
			log.BulletListItem{
				Level:       0,
				Text:        log.Yellow(site),
				TextStyle:   log.NewStyle(log.FgBlue),
				BulletStyle: log.NewStyle(log.FgRed),
			},
		)
	}

	// Print the available scrapers as a bullet list
	if err := log.DefaultBulletList.WithItems(availableScrapers).Render(); err != nil {
		log.Error.Println("Could not print available scrapers", log.Red(err))
		os.Exit(1)
	}

	// Show example of usage
	log.DefaultSection.Println("Usage")
	log.DefaultBasicText.Println("Use the", log.Blue("--website"), "flag like",
		log.Blue("--website"), log.White("blubeaver"), "to start scraping.",
	)

	// Show how to contribute
	log.DefaultSection.Println("Contribution")
	log.DefaultBasicText.Println("See how you can add support for a new website:",
		log.White("https://github.com/kinoute/moviestills#contribute"),
	)

}

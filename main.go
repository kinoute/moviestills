package main

import (
    "log"
    "moviestills/websites"
    "os"
    "strings"
    "time"

    "github.com/alexflint/go-arg"
    "github.com/gocolly/colly/v2"
    "github.com/gocolly/colly/v2/debug"
    "github.com/gocolly/colly/v2/extensions"
)

const VERSION string = "0.1.0"

// Handle arguments passed through the CLI
type args struct {
    Website  string        `arg:"required,-w, --website,env:WEBSITE" help:"Website to scrap movie stills on"`
    Delay    time.Duration `arg:"-t, --time,env:DELAY" help:"Delay in seconds to avoid getting banned" default:"2s"`
    Parallel int           `arg:"-p, --parallel,env:PARALLEL" help:"Limit the maximum parallelism" default:"2"`
    Async    bool          `arg:"-a, --async,env:ASYNC" help:"Enable asynchronus running job"`
    CacheDir string        `arg:"-c, --cache-dir,env:CACHE_DIR" help:"Where to cache scraped websites pages" default:"cache"`
    Debug    bool          `arg:"-d, --debug,env:DEBUG" help:"Enable debugging for Colly, our scraper"`
}

func (args) Description() string {
    return "this program can scrap various websites to get high quality movie snapshots."
}

func (args) Version() string {
    return "moviestills " + VERSION
}

func main() {

    // Implemented scrapers as today
    sites := map[string]func(**colly.Collector){
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

    var args args
    params := arg.MustParse(&args)

    // Check website CLI argument
    if args.Website == "" {
        params.Fail("A website must be set through arguments.")
    }

    // Verify if we have a scrapper for the given website.
    // If we do, "site_func" will now contain a function listed in
    // the sites map that matches a module for this specific
    // website stored in the "websites" folder.
    site_func, site_available := sites[strings.ToLower(args.Website)]
    if !site_available {
        log.SetFlags(0)
        log.Println("We don't have a scraper for this website.")
        log.Println("List of available scrapers:")
        for site := range sites {
            log.Println("–", site)
        }
        os.Exit(1)
    }

    // If we're here, it means we have a valid scraper for a valid website!

    // Create the "cache" directory.
    // This folder stores the scraped websites pages.
    // If we can't create it, stop right there.
    if _, err := os.Stat(args.CacheDir); os.IsNotExist(err) {
        if err = os.Mkdir(args.CacheDir, os.ModePerm); err != nil {
            log.Fatalln("The cache directory", args.CacheDir, "can't be created:", err)
        }
    }

    // Instantiate main scraper
    scraper := colly.NewCollector(
        colly.CacheDir(args.CacheDir),
    )

    // Enable asynchronous jobs if asked
    if args.Async {
        scraper.Async = true
    }

    // Enable Colly Debugging if asked through the CLI
    if args.Debug {
        scraper.SetDebugger(&debug.LogDebugger{})
    }

    // Use random user agent and referer to avoid getting banned
    extensions.RandomUserAgent(scraper)
    extensions.Referer(scraper)

    // Limit parallelism and add random delay to avoid getting IP banned
    if err := scraper.Limit(&colly.LimitRule{
        DomainGlob:  "*",
        Parallelism: args.Parallel,
        RandomDelay: args.Delay * time.Second,
    }); err != nil {
        log.Println("Can't change scraper limit options:", err)
    }

    // Here we call the website module depending on the website provided
    // in the CLI by the user.
    // This will call a file/module/func made specifically to scrap this website.
    // All available scrapers are stored in the "websites" folder.
    site_func(&scraper)

}

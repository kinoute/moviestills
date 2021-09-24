package main

import (
    "log"
    "moviestills/websites"
    "os"
    "strings"
    "time"

    arg "github.com/alexflint/go-arg"
    "github.com/gocolly/colly/v2"
    "github.com/gocolly/colly/v2/extensions"
    "github.com/velebak/colly-sqlite3-storage/colly/sqlite3"
)

func main() {

    // implemented scrapers as today
    sites := map[string]func(**colly.Collector){
        "dvdbeaver":   websites.DVDBeaverScraper,
        "blusscreens": websites.BlusScraper,
    }

    // ask for website to scrap through arguments
    var args struct {
        Website string        `arg:"required,-w, --website" help:"Website to scrap movie stills on"`
        Delay   time.Duration `arg:"-d, --delay" help:"Delay in seconds to avoid getting banned" default:"5s"`
    }

    params := arg.MustParse(&args)

    // check website argument CLI
    if args.Website == "" {
        params.Fail("A website must be set through arguments.")
    }

    // verify if we have a scrapper for the given website
    // if we do, site_func will now be a function listed in
    // the sites map that matches a module for this specifi
    // website.
    site_func, site_available := sites[strings.ToLower(args.Website)]
    if !site_available {
        log.Println("We don't have a scraper for this website.")
        log.Println("List of available scrapers:")
        for site := range sites {
            log.Println("–", site)
        }
        os.Exit(1)
    }

    // if we're here, it means we have a valid scraper
    // for a valid website!

    // Instantiate collector
    scraper := colly.NewCollector(
        colly.CacheDir("./cache"),
    )

    // save state of the scrapping on disk
    storage := &sqlite3.Storage{
        Filename: "./progress.db",
    }

    defer storage.Close()

    err := scraper.SetStorage(storage)
    if err != nil {
        panic(err)
    }

    // use random user agent and referer to
    // avoid getting banned
    extensions.RandomUserAgent(scraper)
    extensions.Referer(scraper)

    // limit parallelism and add rand delay
    // to avoid getting IP banned
    scraper.Limit(&colly.LimitRule{
        Parallelism: 2,
        RandomDelay: args.Delay * time.Second,
    })

    // here we call the website module:
    // depending on the website provided
    // in the CLI, it will call a different
    // file/module/function made specifically
    // to scrap this website.
    site_func(&scraper)
}

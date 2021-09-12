package main

import (
    "log"
    "os"
    "strings"
    "time"

    arg "github.com/alexflint/go-arg"
    "github.com/gocolly/colly/v2"
    "github.com/gocolly/colly/v2/extensions"
    "github.com/velebak/colly-sqlite3-storage/colly/sqlite3"
)

func main() {

    // implemented scraper
    sites := map[string]func(**colly.Collector){
        "dvdbeaver": newBeaverScraper,
    }

    // ask for website to scrap through arguments
    var args struct {
        Website string        `arg:"required,-w, --website" help: "Website to scrap movie stills on"`
        Delay   time.Duration `arg:"-d, --delay" help: "Delay in seconds to avoid getting banned" default:"5"`
    }

    params := arg.MustParse(&args)

    // check website argument CLI
    if args.Website == "" {
        params.Fail("A website must be set through arguments.")
    }

    // verify if we have a scrapper for the given website
    site_func, site_available := sites[strings.ToLower(args.Website)]
    if !site_available {
        log.Println("We don't have a scraper for this website.")
        log.Println("List of available scrapers:")
        for site, _ := range sites {
            log.Println("– ", site)
        }
        os.Exit(1)
    }

    // if we're here, it means we have a valid scraper
    // for a valid website!

    // Instantiate collector
    c := colly.NewCollector(
        colly.CacheDir("./cache"),
    )
    // save state of the scrapping on disk
    storage := &sqlite3.Storage{
        Filename: "./progress.db",
    }

    defer storage.Close()

    err := c.SetStorage(storage)
    if err != nil {
        panic(err)
    }

    // use random user agent and referer to
    // avoid getting banned
    extensions.RandomUserAgent(c)
    extensions.Referer(c)

    // limit parallelism and add rand delay
    // to avoid getting IP banned
    c.Limit(&colly.LimitRule{
        Parallelism: 2,
        RandomDelay: args.Delay * time.Second,
    })

    // here we call the website module
    // depending on the website provided
    // in the CLI, it will call a different
    // file/module/function made specifically
    // to scrap this website.
    site_func(&c)
}

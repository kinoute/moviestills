package websites

import (
	"moviestills/config"
	"moviestills/scraper"
	"moviestills/utils"

	"github.com/gocolly/colly/v2"
	"github.com/pterm/pterm"
)

// EvanERichardsURL is the webpage that stores a list of links to movie,
// TV movies, Series...
const EvanERichardsURL string = "https://www.evanerichards.com/index"

// EvanERichardsScraper is the main function that handles all the scraping logic
// for this website.
func EvanERichardsScraper(c *colly.Collector, options *config.Options, stats *scraper.Stats) {
	log := scraper.NewLogger("evanerichards")

	cfg := scraper.SiteConfig{
		Name:     "evanerichards",
		IndexURL: EvanERichardsURL,
		AllowedDomains: []string{
			"www.evanerichards.com",
			"evanerichards.com",
		},
	}

	// Setup the index scraper with common settings
	scraper.SetupIndexScraper(c, cfg, log)

	// Create and setup the movie scraper
	movieScraper := scraper.SetupMovieScraper(c, log)

	// Setup the common image response handler
	scraper.SetupImageResponseHandler(movieScraper, options, stats, log)

	// Find links to movies pages and isolate the movie's title and year.
	// We iterate through each table row to check if it's indeed a movie
	// and not something else –– this website provides TV Series too.
	c.OnHTML("tbody tr.pp-table-row", func(e *colly.HTMLElement) {
		// Fetch various data in columns for each table entry
		title, _ := utils.Normalize(e.DOM.Find("td.pp-table-cell-Title a").Text())
		category, _ := utils.Normalize(e.DOM.Find("td.pp-table-cell-Category").Text())
		year, _ := utils.Normalize(e.DOM.Find("td.pp-table-cell-Date").Text())

		// Ignore entries that are not movies
		if category != "Movie" && category != "Animation" {
			log.Debug(pterm.White(title), "is not a Movie, ignoring...")
			return
		}

		movieURL, urlExists := e.DOM.Find("td.pp-table-cell-Title a").Attr("href")
		if !urlExists {
			log.Debug("could not find movie URL, next")
			return
		}

		log.Debug("Found movie page link", pterm.White(movieURL))

		movie := scraper.NewMovie(title, year, movieURL, "evanerichards", options)
		log.Info("Found movie page for:", pterm.White(title))

		if stats != nil {
			stats.IncrMovies()
		}

		if err := movieScraper.Request("GET", movieURL, nil, movie.ToContext(), nil); err != nil {
			log.Error("Can't get movie page", pterm.White(movieURL), ":", pterm.Red(err))
		}
	})

	// Look for links on thumbnails that redirect to a "largest" version
	movieScraper.OnHTML("div.elementor-widget-container div.ngg-gallery-thumbnail a[class*=shutter]", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("Found linked image", pterm.White(movieImageURL))

		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error("Can't get large image", pterm.White(movieImageURL), ":", pterm.Red(err))
		}
	})

	// Visit and wait for completion
	scraper.VisitAndWait(c, movieScraper, EvanERichardsURL, log)
}

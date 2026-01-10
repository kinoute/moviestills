package websites

import (
	"moviestills/config"
	"moviestills/scraper"
	"moviestills/utils"

	"github.com/gocolly/colly/v2"
	"github.com/pterm/pterm"
)

// StillsFrmFilmsURL is a webpage that stores a list of links to movies
const StillsFrmFilmsURL string = "https://stillsfrmfilms.wordpress.com/movies-a-z/"

// StillsFrmFilmsScraper is the main function that handles all the scraping logic
// for this website.
func StillsFrmFilmsScraper(c *colly.Collector, options *config.Options, stats *scraper.Stats) {
	log := scraper.NewLogger("stillsfrmfilms")

	cfg := scraper.SiteConfig{
		Name:     "stillsfrmfilms",
		IndexURL: StillsFrmFilmsURL,
		AllowedDomains: []string{
			"stillsfrmfilms.wordpress.com",
			"stillsfrmfilms.files.wordpress.com",
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
	c.OnHTML("div.page-body div.wp-caption", func(e *colly.HTMLElement) {
		// Isolate the movie's title from the description
		movieName, err := utils.Normalize(e.DOM.Find("p.wp-caption-text").Text())
		if err != nil {
			log.Error("Can't normalize the movie title", pterm.Red(err))
			return
		}

		// Isolate the movie page URL
		movieURL, urlExists := e.DOM.Find("a[href*=stills]").Attr("href")
		if !urlExists {
			log.Debug("Can't find URL to movie page, next")
			return
		}

		log.Debug("Found movie page link", pterm.White(movieURL))

		movie := scraper.NewMovie(movieName, "", movieURL, "stillsfrmfilms", options)
		log.Info("Found movie page for:", pterm.White(movieName))

		if stats != nil {
			stats.IncrMovies()
		}

		if err := movieScraper.Request("GET", movieURL, nil, movie.ToContext(), nil); err != nil {
			log.Error("Can't get movie page", pterm.White(movieURL), ":", pterm.Red(err))
		}
	})

	// Look for links on thumbnails that redirect to a "largest" version.
	movieScraper.OnHTML("div.photo-inner dl.gallery-item a[href*=stills] img[src*=uploads]", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("data-orig-file"))

		// Use regexp to remove potential GET parameters from the URL
		// regarding the resolution of the displayed image.
		movieImageURL = utils.RemoveURLParams(movieImageURL)

		log.Debug("Found linked image", pterm.White(movieImageURL))
		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error("Can't get movie image", pterm.White(movieImageURL), ":", pterm.Red(err))
		}
	})

	// Visit and wait for completion
	scraper.VisitAndWait(c, movieScraper, StillsFrmFilmsURL, log)
}

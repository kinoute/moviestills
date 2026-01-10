package websites

import (
	"moviestills/config"
	"moviestills/scraper"
	"moviestills/utils"

	"github.com/gocolly/colly/v2"
	"github.com/pterm/pterm"
)

// FilmGrabURL is the webpage that stores a list of links to movie pages
// sorted by alphabet.
const FilmGrabURL string = "https://film-grab.com/movies-a-z/"

// FilmGrabScraper is the main function that handles all the scraping
// logic for this website.
func FilmGrabScraper(c *colly.Collector, options *config.Options, stats *scraper.Stats) {
	log := scraper.NewLogger("film-grab")

	cfg := scraper.SiteConfig{
		Name:     "film-grab",
		IndexURL: FilmGrabURL,
		AllowedDomains: []string{
			"film-grab.com",
		},
	}

	// Setup the index scraper with common settings
	scraper.SetupIndexScraper(c, cfg, log)

	// Create and setup the movie scraper
	movieScraper := scraper.SetupMovieScraper(c, log)

	// Setup the common image response handler
	scraper.SetupImageResponseHandler(movieScraper, options, stats, log)

	// Find links to movies pages and isolate the movie's title.
	c.OnHTML("div#primary a.title[href*=film]", func(e *colly.HTMLElement) {
		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("Found movie page link", pterm.White(movieURL))

		// Remove weird accents and spaces from the movie's title
		movieName, err := utils.Normalize(e.Text)
		if err != nil {
			log.Error("Can't normalize Movie name for", pterm.White(e.Text), pterm.Red(err))
			return
		}

		movie := scraper.NewMovie(movieName, "", movieURL, "film-grab", options)
		log.Info("Found movie page for:", pterm.White(movieName))

		if stats != nil {
			stats.IncrMovies()
		}

		if err := movieScraper.Request("GET", movieURL, nil, movie.ToContext(), nil); err != nil {
			log.Error("Can't get movie page", pterm.White(movieURL), ":", pterm.Red(err))
		}
	})

	// Look for links on thumbnails that redirect to a "largest" version
	movieScraper.OnHTML("div.bwg_container div.bwg-item a.bwg-a[href*=film]", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))

		// Remove weird GET parameters to have a proper filename
		movieImageURL = utils.RemoveURLParams(movieImageURL)

		log.Debug("Found link to large image", pterm.White(movieImageURL))

		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error("Can't request linked image:", pterm.Red(err))
		}
	})

	// Visit and wait for completion
	scraper.VisitAndWait(c, movieScraper, FilmGrabURL, log)
}

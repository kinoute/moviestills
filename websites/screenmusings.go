package websites

import (
	"moviestills/config"
	"moviestills/scraper"
	"moviestills/utils"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/pterm/pterm"
)

// ScreenMusingsURL is the page that lists all movies available, sorted alphabetically
const ScreenMusingsURL string = "https://screenmusings.org/movie/"

// ScreenMusingsScraper is the main function that handles all the scraping logic
// for this website.
func ScreenMusingsScraper(c *colly.Collector, options *config.Options, stats *scraper.Stats) {
	log := scraper.NewLogger("screenmusings")

	cfg := scraper.SiteConfig{
		Name:     "screenmusings",
		IndexURL: ScreenMusingsURL,
		AllowedDomains: []string{
			"screenmusings.org",
		},
	}

	// Setup the index scraper with common settings
	scraper.SetupIndexScraper(c, cfg, log)

	// Create and setup the movie scraper
	movieScraper := scraper.SetupMovieScraper(c, log)

	// Setup the common image response handler
	scraper.SetupImageResponseHandler(movieScraper, options, stats, log)

	// Isolate every movie listed, keep its title and year.
	// Create a dedicated folder if it doesn't exist to store images.
	//
	// Then visit movie page where images are listed/displayed. It seems
	// this website has both DVD and Blu-Rays reviews, let's take care of it.
	c.OnHTML("nav#movies ul li a[href*=dvd], nav#movies ul li a[href*=blu]", func(e *colly.HTMLElement) {
		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("Found movie page link", pterm.White(movieURL))

		// Take care of weird accents and spaces
		movieName, err := utils.Normalize(e.Text)
		if err != nil {
			log.Error("Can't normalize Movie name for", pterm.White(e.Text), ":", pterm.Red(err))
			return
		}

		movie := scraper.NewMovie(movieName, "", movieURL, "screenmusings", options)
		log.Info("Found movie page for:", pterm.White(movieName))

		if stats != nil {
			stats.IncrMovies()
		}

		if err = movieScraper.Request("GET", movieURL, nil, movie.ToContext(), nil); err != nil {
			log.Error("Can't get movie page", pterm.White(movieURL), ":", pterm.Red(err))
		}
	})

	// On every movie page, we are looking for a link to the "most viewed stills".
	// This link is extremely handy as it seems to display every thumbnail on a
	// single page. Therefore, we don't have to deal with pagination.
	movieScraper.OnHTML("ul#gallery-nav-top li:nth-last-child(2) a[href*=most]", func(e *colly.HTMLElement) {
		mostViewedImages := e.Attr("href")
		log.Debug("get most viewed stills link for", pterm.White(e.Request.Ctx.Get("movie_name")))
		if err := e.Request.Visit(mostViewedImages); err != nil {
			log.Error("Can't request most viewed stills page:", pterm.Red(err))
		}
	})

	// We iterate through every thumbnail on the "most viewed stills" page.
	// We have to replace "thumbnails" in the URL by "images" to get
	// the URL that links to the full resolution image.
	movieScraper.OnHTML("div#thumbnails div.thumb img[src*=thumb]", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("src"))

		// Replace "thumbnails" by "images" to get the full image URL
		movieImageURL = strings.Replace(movieImageURL, "thumbnails", "images", 1)

		log.Debug("Found linked image", pterm.White(movieImageURL))
		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error("Can't request linked image", pterm.White(movieImageURL), pterm.Red(err))
		}
	})

	// Visit and wait for completion
	scraper.VisitAndWait(c, movieScraper, ScreenMusingsURL, log)
}

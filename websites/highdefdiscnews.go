package websites

import (
	"moviestills/config"
	"moviestills/scraper"
	"moviestills/utils"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/pterm/pterm"
)

// HighDefDiscNewsURL is the webpage that stores a list of links to movie pages
// with Blu-rays images.
const HighDefDiscNewsURL string = "https://highdefdiscnews.com/blu-ray-screenshots/"

// HighDefDiscNewsScraper is the main function that handles all the scraping logic
// for this website.
func HighDefDiscNewsScraper(c *colly.Collector, options *config.Options, stats *scraper.Stats) {
	log := scraper.NewLogger("highdefdiscnews")

	cfg := scraper.SiteConfig{
		Name:     "highdefdiscnews",
		IndexURL: HighDefDiscNewsURL,
		AllowedDomains: []string{
			"highdefdiscnews.com",
		},
	}

	// Setup the index scraper with common settings
	scraper.SetupIndexScraper(c, cfg, log)

	// Create and setup the movie scraper
	movieScraper := scraper.SetupMovieScraper(c, log)

	// Setup the common image response handler
	scraper.SetupImageResponseHandler(movieScraper, options, stats, log)

	// Find links to movies reviews and isolate the movie's title.
	// Links contain some useless text such as "- Blu-ray Screenshots"
	// or "[Remastered]".
	// We remove these texts to isolate the movie's title.
	c.OnHTML("div#mcTagMap ul.links a[href*=high]", func(e *colly.HTMLElement) {
		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("Found movie page link", pterm.White(movieURL))

		// Isolate the movie's title from the text
		tmpMovieName := isolateMovieTitle(e.Text)

		// Remove weird accents and spaces from the movie's title
		movieName, err := utils.Normalize(tmpMovieName)
		if err != nil {
			log.Error("Can't normalize Movie name for", pterm.White(e.Text), pterm.Red(err))
			return
		}

		movie := scraper.NewMovie(movieName, "", movieURL, "highdefdiscnews", options)
		log.Info("Found movie page for:", pterm.White(movieName))

		if stats != nil {
			stats.IncrMovies()
		}

		if err := movieScraper.Request("GET", movieURL, nil, movie.ToContext(), nil); err != nil {
			log.Error("Can't visit movie page", pterm.White(movieURL), ":", pterm.Red(err))
		}
	})

	// Look for links on thumbnails that redirects to a "largest" version.
	movieScraper.OnHTML("div.gallery dl.gallery-item a[href*=high]", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("Found linked image", pterm.White(movieImageURL))
		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error("Can't get linked image", pterm.White(movieImageURL), ":", pterm.Red(err))
		}
	})

	// Visit and wait for completion
	scraper.VisitAndWait(c, movieScraper, HighDefDiscNewsURL, log)
}

// isolateMovieTitle isolates the movie's title by getting rid of various words on the right.
// eg. "[Remastered]", "- Blu-ray Screenshots".
func isolateMovieTitle(sentence string) string {
	// Get rid of text such as "- Blu-ray Screenshots"
	if idx := strings.LastIndex(sentence, " - "); idx != -1 {
		sentence = sentence[:idx]
	}

	// Get rid of text such as "[Remastered]"
	if idx := strings.LastIndex(sentence, " ["); idx != -1 {
		sentence = sentence[:idx]
	}

	return sentence
}

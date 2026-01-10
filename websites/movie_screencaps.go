package websites

import (
	"moviestills/config"
	"moviestills/scraper"
	"moviestills/utils"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/pterm/pterm"
)

// ScreenCapsURL is the page that lists all movies available, sorted alphabetically
const ScreenCapsURL string = "https://movie-screencaps.com/movie-directory/"

// ScreenCapsScraper is the main function that handles all the scraping logic
// for this website.
func ScreenCapsScraper(c *colly.Collector, options *config.Options, stats *scraper.Stats) {
	log := scraper.NewLogger("movie-screencaps")

	cfg := scraper.SiteConfig{
		Name:     "movie-screencaps",
		IndexURL: ScreenCapsURL,
		AllowedDomains: []string{
			"movie-screencaps.com",
			"www.movie-screencaps.com",
			"i0.wp.com",
			"i1.wp.com",
			"i2.wp.com",
			"i3.wp.com",
			"wp.com",
			"img.screencaps.us",
		},
	}

	// Setup the index scraper with common settings
	scraper.SetupIndexScraper(c, cfg, log)

	// Create and setup the movie scraper
	movieScraper := scraper.SetupMovieScraper(c, log)

	// Setup the common image response handler
	scraper.SetupImageResponseHandler(movieScraper, options, stats, log)

	// Isolate every movie listed, keep its title and
	// create a dedicated folder if it doesn't exist
	// to store images.
	//
	// Then visit movie page where images are listed/displayed.
	c.OnHTML("div.tagindex ul.links li a[href*=movie]", func(e *colly.HTMLElement) {
		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("Found movie page link", pterm.White(movieURL))

		// Take care of weird accents and spaces
		movieName, err := utils.Normalize(e.Text)
		if err != nil {
			log.Error("Can't normalize Movie name for", pterm.White(e.Text), ":", pterm.Red(err))
			return
		}

		log.Debug("Found movie link for", pterm.White(movieName))

		movie := scraper.NewMovie(movieName, "", movieURL, "movie-screencaps", options)
		log.Info("Found movie page for:", pterm.White(movieName))

		if stats != nil {
			stats.IncrMovies()
		}

		if err = movieScraper.Request("GET", movieURL, nil, movie.ToContext(), nil); err != nil {
			log.Error("Can't get movie page", pterm.White(movieURL), ":", pterm.Red(err))
		}
	})

	// Handle pagination by getting the number of pages in total first.
	// Then iterate through all pages with a for loop to get movie stills.
	movieScraper.OnHTML("div.pixcode + div.wp-pagenavi > select.paginate option:last-child", func(e *colly.HTMLElement) {
		// Get the URL of the movie page
		actualPageURL := e.Request.URL.String()

		// Only start to visit paginated pages if we're at the first page.
		// Otherwise it will result in an infinite loop.
		if !strings.Contains(actualPageURL, "/page") {
			movieName := e.Request.Ctx.Get("movie_name")

			// Get the total number of pages from the select menu and the last option
			numOfPages, _ := strconv.Atoi(e.Attr("value"))
			log.Info("number of pages for", pterm.White(movieName), "is", pterm.White(e.Attr("value")))

			// Visit every paginated page to get a few snapshots every time
			for num := 2; num <= numOfPages; num++ {
				log.Info("visiting paginated page", pterm.White(strconv.Itoa(num)), "for", pterm.White(movieName))
				paginatedPageURL := actualPageURL + "page/" + strconv.Itoa(num)
				if err := e.Request.Visit(paginatedPageURL); err != nil {
					log.Error("Can't visit paginated page", pterm.White(paginatedPageURL), ":", pterm.Red(err))
				}
			}
		}
	})

	// Go through each link to a movie snapshot found on the movie page.
	//
	// Exceptionally, since this website basically takes a snapshot every
	// second or so during the movie, if we download everything, we will have
	// many similar snapshots and it's going to take forever.
	//
	// Therefore, we added :nth-of-type(30n) to the CSS selector to only
	// download 1 shot every 30 shots. Remove it if you want to download everything.
	movieScraper.OnHTML("section.entry-content a[href*=wp][href*=caps]:nth-of-type(30n)", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))

		// We're getting weird filenames from Wordpress with "strip=all" at the end.
		// We might need to remove some suffixes.
		movieImageURL = utils.RemoveURLParams(movieImageURL)

		log.Debug("Found linked image", pterm.White(movieImageURL))
		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error("Can't request linked image", pterm.White(movieImageURL), pterm.Red(err))
		}
	})

	movieScraper.OnHTML("section.entry-content a[href*=screencaps][href$=jpg]:nth-of-type(30n)", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))

		// We're getting weird filenames from Wordpress with "strip=all" at the end.
		// We might need to remove some suffixes.
		movieImageURL = utils.RemoveURLParams(movieImageURL)

		log.Debug("Found linked image", pterm.White(movieImageURL))
		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error("Can't request linked image", pterm.White(movieImageURL), pterm.Red(err))
		}
	})

	// Visit and wait for completion
	scraper.VisitAndWait(c, movieScraper, ScreenCapsURL, log)
}

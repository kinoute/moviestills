package websites

import (
	"moviestills/config"
	"moviestills/scraper"
	"moviestills/utils"
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/pterm/pterm"
)

// BluBeaverURL is the webpage stores a list of links to movie reviews of Blu-rays
const BluBeaverURL string = "http://www.dvdbeaver.com/blu-ray.htm"

// BluBeaverScraper is the main function that handles all the scraping logic
// for this website.
func BluBeaverScraper(c *colly.Collector, options *config.Options, stats *scraper.Stats) {
	log := scraper.NewLogger("blubeaver")

	cfg := scraper.SiteConfig{
		Name:     "blubeaver",
		IndexURL: BluBeaverURL,
		AllowedDomains: []string{
			"www.blubeaver.ca",
			"www.dvdbeaver.com",
			"dvdbeaver.com",
			"DVDBeaver.com",
			"www.DVDBeaver.com",
		},
		DetectCharset: true,
	}

	// Setup the index scraper with common settings
	scraper.SetupIndexScraper(c, cfg, log)

	// Create and setup the movie scraper
	movieScraper := scraper.SetupMovieScraper(c, log)

	// Setup the common image response handler
	scraper.SetupImageResponseHandler(movieScraper, options, stats, log)

	// Find links to movies reviews and isolate the movie's title.
	// Since BluBeaver is somewhat a custom website, some links
	// might have different cases. We use the CSS4 "i" case-insensitive
	// feature to make sure our filter doesn't miss anything.
	c.OnHTML("li a[href*='film' i][href$='htm' i]", func(e *colly.HTMLElement) {
		// Sometimes, Blubeaver made mistakes and added links to reviews
		// on Amazon icons. Since we use the link to isolate the movie's title,
		// we ignore these links as they don't have the movie's name included.
		if _, iconExistsInLink := e.DOM.Find("img").Attr("src"); iconExistsInLink {
			log.Debug("Link without text, just an icon, next")
			return
		}

		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("Found movie page link", pterm.White(movieURL))

		// Remove weird accents and spaces from the movie's title
		movieName, err := utils.Normalize(e.Text)
		if err != nil {
			log.Error("Can't normalize Movie name for", pterm.White(e.Text), ":", pterm.Red(err))
			return
		}

		movie := scraper.NewMovie(movieName, "", movieURL, "blubeaver", options)
		log.Info("Found movie page for:", pterm.White(movieName))

		if stats != nil {
			stats.IncrMovies()
		}

		if err = movieScraper.Request("GET", movieURL, nil, movie.ToContext(), nil); err != nil {
			log.Error("Can't get movie page", pterm.White(movieURL), ":", pterm.Red(err))
		}
	})

	// It's rare but sometimes on BD reviews there are no large versions.
	// Therefore we download the images as shown on the webpage and
	// be sure we avoid some weird ones (subtitles, DVD covers etc).
	movieScraper.OnHTML(
		":not(a) >"+
			"img:not([src*='banner' i])"+
			":not([src*='rating' i])"+
			":not([src*='package' i])"+
			":not([src*='bitrate' i])"+
			":not([src*='bitgraph' i])"+
			":not([src$='gif' i])"+
			":not([src*='sub' i])"+
			":not([src*='daggers' i])"+
			":not([src*='poster' i])"+
			":not([src*='title' i])"+
			":not([src*='menu' i])", func(e *colly.HTMLElement) {
			movieImageURL := e.Request.AbsoluteURL(e.Attr("src"))

			// Filter low resolutions images to avoid false positives.
			// if the images are too small, we won't be able to use them
			// anyway so let's skip them.
			movieImageWidth, _ := strconv.Atoi(e.Attr("width"))
			movieImageHeight, _ := strconv.Atoi(e.Attr("height"))

			if movieImageHeight >= 265 && movieImageWidth >= 500 {
				if err := e.Request.Visit(movieImageURL); err != nil {
					log.Error("Can't get inline image", pterm.White(movieImageURL), ":", pterm.Red(err))
				}
			}
		})

	// Look for links on images that redirects to a "largest" version.
	// These links appear on Blu-Ray reviews almost exclusively and
	// provide images with native resolution (1080p).
	//
	// We try to avoid images with "subs" in the filename as they are
	// most likely images with subtitles on top. We don't want that.
	movieScraper.OnHTML("a[href*='large' i]:not([href*='subs' i])", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("Found large image", pterm.White(movieImageURL))

		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error("Can't get large image", pterm.White(movieImageURL), ":", pterm.Red(err))

			// Sometimes, the high quality version of an image
			// is not available anymore ("Not Found").
			//
			// In this case, we can try to save the image
			// shown on the webpage that has a lower resolution.
			lowImageURL, imgExists := e.DOM.Find("img").Attr("src")
			if !imgExists {
				log.Error("Could not find an image inside link", pterm.White(movieImageURL))
				return
			}

			log.Info("Trying to save low quality image instead", pterm.White(lowImageURL))
			if err := e.Request.Visit(e.Request.AbsoluteURL(lowImageURL)); err != nil {
				log.Error("Can't get low resolution image", pterm.White(lowImageURL), ":", pterm.Red(err))
			}
		}
	})

	// Visit and wait for completion
	scraper.VisitAndWait(c, movieScraper, BluBeaverURL, log)
}

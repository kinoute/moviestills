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

// BeaverURL is the webpage that stores a list of links to movie list pages
// sorted by alphabet (#, a, z). It's a good starting point for our task.
const BeaverURL string = "http://www.dvdbeaver.com/film/reviews.htm"

// DVDBeaverScraper is the main function that handles all the scraping
// logic for this website.
func DVDBeaverScraper(c *colly.Collector, options *config.Options, stats *scraper.Stats) {
	log := scraper.NewLogger("dvdbeaver")

	cfg := scraper.SiteConfig{
		Name:     "dvdbeaver",
		IndexURL: BeaverURL,
		AllowedDomains: []string{
			"www.dvdbeaver.com",
			"DVDBeaver.com",
			"www.DVDBeaver.com",
		},
	}

	// Setup the index scraper with common settings
	scraper.SetupIndexScraper(c, cfg, log)

	// Movies list might be updated often with new movies
	// so we authorize the scraper to revisit these pages.
	movieListScraper := c.Clone()
	movieListScraper.AllowURLRevisit = true
	movieListScraper.DetectCharset = true

	movieListScraper.OnRequest(func(r *colly.Request) {
		log.Debug("visiting movie list page", pterm.White(r.URL.String()))
	})

	// Create and setup the movie scraper
	movieScraper := scraper.SetupMovieScraper(c, log)

	// Setup the common image response handler
	scraper.SetupImageResponseHandler(movieScraper, options, stats, log)

	// Find links to movies list by alphabet
	c.OnHTML("a[href*='listing' i]", func(e *colly.HTMLElement) {
		movieListURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("Found movie list page link", pterm.White(movieListURL))

		if err := movieListScraper.Visit(movieListURL); err != nil {
			log.Error("Can't visit movie list page", pterm.White(movieListURL), pterm.Red(err))
		}
	})

	// Look for movie reviews links and create folder for every
	// movie we find to prepare the download of the snapshots.
	//
	// We have to iterate through each p element to discard
	// Blu-Ray reviews as we want to focus only on DVD reviews.
	movieListScraper.OnHTML("td p", func(e *colly.HTMLElement) {
		// We ignore BD reviews pages because we have
		// a specific scraper, "BluBeaver", for these
		// pages with Blu-Ray screenshots.
		if strings.Contains(e.DOM.Text(), "BD") || strings.Contains(e.DOM.Text(), "UHD") {
			log.Debug("BD review, skipping")
			return
		}

		// Get the DVD movie review link. Sometimes there is
		// no link that matches our query so we stop right here.
		//
		// We use the CSS4 case-insensitive feature "i" to make sure
		// our filter will find everything, no matter the case.
		reviewLink := e.DOM.Find("a[href*='film' i]")
		movieURL, urlExists := reviewLink.Attr("href")
		if !urlExists {
			log.Debug("no movie review link could be found, next")
			return
		}

		// Take care of weird characters in the movie's title
		movieName, err := utils.Normalize(reviewLink.Text())
		if err != nil {
			log.Error("Can't normalize Movie name for", pterm.White(reviewLink.Text()), pterm.Red(err))
			return
		}

		log.Debug("Found movie link for", pterm.White(movieName))

		// Make sure we handle relative URLs if any
		movieURL = e.Request.AbsoluteURL(movieURL)

		movie := scraper.NewMovie(movieName, "", movieURL, "dvdbeaver", options)
		log.Info("Found movie page for:", pterm.White(movieName))

		if stats != nil {
			stats.IncrMovies()
		}

		if err = movieScraper.Request("GET", movieURL, nil, movie.ToContext(), nil); err != nil {
			log.Error("Can't get movie page", pterm.White(movieURL), ":", pterm.Red(err))
		}
	})

	// Look for links on images that redirects to a "largest" version.
	// It is unlikely to find some of these on some DVD reviews, but sometimes
	// they compare DVD releases with BD releases and provide some images
	// with native resolution (1080p).
	//
	// We try to avoid images with "subs" in the filename as they are
	// most likely images with subtitles on top. We don't want that.
	movieScraper.OnHTML("a[href*='large' i]:not([href*='subs' i])", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("Found large image", pterm.White(movieImageURL))

		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error("Can't get large image", pterm.White(movieImageURL), ":", pterm.Red(err))
		}
	})

	// On DVD reviews, there are almost never clickable large versions.
	// Therefore we download the images as shown on the webpage and
	// be sure we avoid some weird ones (subtitles, DVD covers etc).
	movieScraper.OnHTML(
		"img:not([src*='banner' i])"+
			":not([src*='rating' i])"+
			":not([src*='package' i])"+
			":not([src*='bitrate' i])"+
			":not([src*='bitgraph' i])"+
			":not([src$='gif' i])"+
			":not([src$='click.jpg' i])"+
			":not([src$='large_apocalypse.jpg' i])"+
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
					log.Error("Can't request inline image", pterm.White(movieImageURL), pterm.Red(err))
				}
			}
		})

	// Visit and wait for completion
	if err := c.Visit(BeaverURL); err != nil {
		log.Error("Can't visit index page:", pterm.Red(err))
	}

	c.Wait()
	movieListScraper.Wait()
	movieScraper.Wait()
}

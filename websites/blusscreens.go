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

// BlusURL is the cinematographers page URL. We use this
// because the movie's title can be easily scrapped/parsed
// there instead of the screen captures page (#, A, Z) where links are images.
//
// The downside is, some movies (animation?) might be missing
// because they don't have cinematographers associated to it.
const BlusURL string = "https://www.bluscreens.net/cinematographers.html"

// MinimumSize is the threshold in bytes to decide if we want to download
// an image or not.
// It is helpful for this website since most of the images are hosted
// on imgur and some might have been deleted. When an image is deleted
// on imgur, it returns a small image with some text on it. We don't want that.
const MinimumSize int = 1024 * 20

// BlusScraper is the main function that handles all the scraping
// logic for this website.
func BlusScraper(c *colly.Collector, options *config.Options, stats *scraper.Stats) {
	log := scraper.NewLogger("blusscreens")

	cfg := scraper.SiteConfig{
		Name:     "blusscreens",
		IndexURL: BlusURL,
		AllowedDomains: []string{
			"www.bluscreens.net",
			"imgur.com",
			"i.imgur.com",
			"postimage.org",
			"postimg.cc",
			"i.postimg.cc",
			"pixxxels.cc",
			"i.pixxxels.cc",
		},
	}

	// Setup the index scraper with common settings
	scraper.SetupIndexScraper(c, cfg, log)

	// Create and setup the movie scraper
	movieScraper := scraper.SetupMovieScraper(c, log)

	// Custom response handler for this site (needs size filtering)
	movieScraper.OnResponse(func(r *colly.Response) {
		// Ignore anything that is not an image
		if !strings.Contains(r.Headers.Get("Content-Type"), "image") {
			return
		}

		// Calculate Image Size from Headers
		imageSize, err := strconv.Atoi(r.Headers.Get("Content-Length"))
		if err != nil {
			log.Error("Can't get image size from headers:", pterm.Red(err))
			return
		}

		// Images are hosted on imgur and some might have been deleted. When an image is deleted
		// on imgur, it returns a small image with some text on it. We don't want that.
		if imageSize < MinimumSize {
			log.Error("Small-sized image, not downloading", pterm.White(r.FileName()))
			if stats != nil {
				stats.IncrFailed()
			}
			return
		}

		movie := scraper.MovieFromContext(r.Ctx)

		if err := scraper.SaveImage(movie.Path, movie.Name, r.FileName(), r.Body, options.Hash, log); err != nil {
			log.Error("Can't save image", pterm.White(r.FileName()), pterm.Red(err))
			if stats != nil {
				stats.IncrFailed()
			}
			return
		}

		if stats != nil {
			stats.IncrDownloaded()
		}
	})

	// Isolate every movie listed, keep its title and
	// create a dedicated folder if it doesn't exist
	// to store images.
	//
	// Then visit movie page where images are listed/displayed.
	c.OnHTML("h2.wsite-content-title a[href*=html]", func(e *colly.HTMLElement) {
		// Remove weird accents and spaces from the movie's title
		movieName, err := utils.Normalize(e.Text)
		if err != nil {
			log.Error("Can't normalize Movie name for", pterm.White(e.Text), ":", pterm.Red(err))
			return
		}

		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("Found movie page link", pterm.White(movieURL))

		movie := scraper.NewMovie(movieName, "", movieURL, "blusscreens", options)
		log.Info("Found movie page for:", pterm.White(movieName))

		if stats != nil {
			stats.IncrMovies()
		}

		if err = movieScraper.Request("GET", movieURL, nil, movie.ToContext(), nil); err != nil {
			log.Error("Can't get movie page", pterm.White(movieURL), ":", pterm.Red(err))
		}
	})

	// Go through each link to imgur found on the movie page
	movieScraper.OnHTML(
		"div.galleryInnerImageHolder a[href*=imgur], "+
			"td.wsite-multicol-col div a[href*=imgur]", func(e *colly.HTMLElement) {
			movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
			log.Debug("inside movie page for", pterm.White(e.Request.Ctx.Get("movie_name")))

			// Create link to the real image if it's a link to imgur's
			// website and not directly to the image.
			// eg. https://imgur.com/ABC to https://i.imgur.com/ABC.png
			if !strings.Contains(movieImageURL, "i.imgur.com") {
				movieImageURL += ".png"
				movieImageURL = strings.Replace(movieImageURL, "https://imgur.com", "https://i.imgur.com", 1)
			}

			log.Debug("Found linked image", pterm.White(movieImageURL))
			if err := e.Request.Visit(movieImageURL); err != nil {
				log.Error("Can't get linked image", pterm.White(movieImageURL), ":", pterm.Red(err))
			}
		})

	// Some old pages of blusscreens have a different layout.
	// We need a special function to handle this.
	// eg: https://www.bluscreens.net/skin-i-live-in-the.html
	movieScraper.OnHTML("div.galleryInnerImageHolder a[href*=postimage]", func(e *colly.HTMLElement) {
		postImgURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("found postimage link", pterm.White(postImgURL))

		if err := e.Request.Visit(postImgURL); err != nil {
			log.Error("Can't request postimage link", pterm.White(postImgURL), ":", pterm.Red(err))
		}
	})

	// Another kind of weird layout mixing table and div.
	// eg: https://www.bluscreens.net/pain--gain.html
	movieScraper.OnHTML("td.wsite-multicol-col div a[href*=postim]", func(e *colly.HTMLElement) {
		postImgURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug("found postimage.org link", pterm.White(postImgURL))

		// Some links redirect to "postimg.org" and later "pixxxels.cc".
		// "postimg.org" is not available anymore, we might need to rewrite the URLs.
		postImgURL = strings.Replace(postImgURL, "postimg.org", "postimage.org", 1)

		if err := e.Request.Visit(postImgURL); err != nil {
			log.Error("Can't request postimage link", pterm.White(postImgURL), ":", pterm.Red(err))
		}
	})

	// Get full images from postimage.cc host.
	// We need to get the "download" button link as
	// the image shown on the page is in a "lower" resolution.
	movieScraper.OnHTML(
		"div#content a#download[href*=postimg], "+
			"div#content a#download[href*=pixxxels]", func(e *colly.HTMLElement) {
			movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
			log.Debug("found postimg full image", pterm.White(movieImageURL))

			if err := e.Request.Visit(movieImageURL); err != nil {
				log.Error("Can't get postimage full image", pterm.White(movieImageURL), ":", pterm.Red(err))
			}
		})

	// Visit and wait for completion
	scraper.VisitAndWait(c, movieScraper, BlusURL, log)
}

package websites

import (
	"moviestills/config"
	"moviestills/utils"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/pterm/pterm"

	"github.com/gocolly/colly/v2"
)

// BeaverURL is the webpage that stores a list of links to movie list pages
// sorted by alphabet (#, a, z). It's a good starting point for our task.
const BeaverURL string = "http://www.dvdbeaver.com/film/reviews.htm"

// DVDBeaverScraper is the main function that handles all the scraping
// logic for this website.
func DVDBeaverScraper(scraper **colly.Collector, options *config.Options) {

	// Change allowed domain for the main scraper.
	// Since everything is served on the same domain, only one domain is necessary.
	(*scraper).AllowedDomains = []string{
		"www.dvdbeaver.com",
		"DVDBeaver.com",
		"www.DVDBeaver.com",
	}

	// The reviews page might have been updated so
	// we have to revisit it when restarting the scraper.
	// It is a single page, it will not cost anything anyway.
	(*scraper).AllowURLRevisit = true

	// Movies list might be updated often with new movies
	// so we autorize the scraper to revisit these pages.
	movieListScraper := (*scraper).Clone()
	movieListScraper.AllowURLRevisit = true

	// DVDBeaver pages have a weird charset.
	// Colly can deal with this automatically
	// and handle weird accents/characters better.
	movieListScraper.DetectCharset = true

	// Scraper to fetch movie images on reviews pages.
	// These pages are not updated after being
	// published therefore we only visit them once.
	movieScraper := (*scraper).Clone()
	movieScraper.AllowURLRevisit = false

	// Print error just in case
	(*scraper).OnError(func(r *colly.Response, err error) {
		log.Error.Println(r.Request.URL, "\t", log.White(r.StatusCode), "\nError:", log.Red(err))
	})

	// Before making a request print "Visiting ..."
	(*scraper).OnRequest(func(r *colly.Request) {
		log.Debug.Println("visiting index page", log.White(r.URL.String()))
	})

	// Find links to movies list by alphabet
	(*scraper).OnHTML("a[href*='listing' i]", func(e *colly.HTMLElement) {
		movieListURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug.Println("Found movie list page link", log.White(movieListURL))

		if err := movieListScraper.Visit(movieListURL); err != nil {
			log.Error.Println("Can't visit movie list page", log.White(movieListURL), log.Red(err))
		}
	})

	// Before making a request print "Visiting ..."
	movieListScraper.OnRequest(func(r *colly.Request) {
		log.Debug.Println("visiting movie list page", log.White(r.URL.String()))
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
			log.Debug.Println("BD review, skipping")
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
			log.Debug.Println("no movie review link could be found, next")
			return
		}

		// Take care of weird characters in the movie's title
		movieName, err := utils.Normalize(reviewLink.Text())
		if err != nil {
			log.Error.Println("Can't normalize Movie name for", log.White(reviewLink.Text()), log.Red(err))
			return
		}

		moviePath := filepath.Join(options.DataDir, options.Website, movieName)
		log.Debug.Println("Found movie link for", log.White(movieName))

		// Pass the movie path to the next request context
		// in order to save the images in the correct folder.
		ctx := colly.NewContext()
		ctx.Put("movie_path", moviePath)
		ctx.Put("movie_name", movieName)

		// Make sure we handle relative URLs if any
		movieURL = e.Request.AbsoluteURL(movieURL)
		log.Info.Println("Found movie page for:", log.White(movieName))

		if err = movieScraper.Request("GET", movieURL, nil, ctx, nil); err != nil {
			log.Error.Println("Can't get movie page", log.White(movieURL), ":", log.Red(err))
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
		log.Debug.Println("Found large image", log.White(movieImageURL))

		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error.Println("Can't get large image", log.White(movieImageURL), ":", log.Red(err))
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
					log.Error.Println("Can't request inline image", log.White(movieImageURL), log.Red(err))
				}
			}
		})

	// Before making a request to URL
	movieScraper.OnRequest(func(r *colly.Request) {
		log.Debug.Println("visiting", log.White(r.URL.String()))
	})

	// Check if what we just visited is an image and
	// save it to the movie folder we created earlier.
	movieScraper.OnResponse(func(r *colly.Response) {

		// Ignore anything that is not an image
		if !strings.Contains(r.Headers.Get("Content-Type"), "image") {
			return
		}

		// Try to save movie image
		if err := utils.SaveImage(r.Ctx.Get("movie_path"),
			r.Ctx.Get("movie_name"),
			r.FileName(),
			r.Body,
			options.Hash,
		); err != nil {
			log.Error.Println("Can't save image", log.White(r.FileName()), log.Red(err))
		}
	})

	if err := (*scraper).Visit(BeaverURL); err != nil {
		log.Error.Println("Can't visit index page:", log.Red(err))
	}

	// Ensure that all requests are completed before exiting
	(*scraper).Wait()
	movieScraper.Wait()
}

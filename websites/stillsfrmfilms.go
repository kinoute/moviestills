package websites

import (
	"moviestills/config"
	"moviestills/utils"
	"strings"

	log "github.com/pterm/pterm"

	"github.com/gocolly/colly/v2"
)

// This webpage stores a list of links to movies
const StillsFrmFilmsURL string = "https://stillsfrmfilms.wordpress.com/movies-a-z/"

// Main function that handles all the scraping logic for this website
func StillsFrmFilmsScraper(scraper **colly.Collector, options *config.Options) {

	// Change allowed domain for the main scraper.
	// Since everything is served on the same domain,
	// only one domain is really necessary.
	// We *do* need to add a subdomain that hosts the images.
	(*scraper).AllowedDomains = []string{
		"stillsfrmfilms.wordpress.com",
		"stillsfrmfilms.files.wordpress.com",
	}

	// The index page might have been updated since last visit so
	// we have to revisit it when restarting the scraper.
	// It is a single page, it will not cost anything anyway.
	(*scraper).AllowURLRevisit = true

	// Scraper to fetch movie images on movie pages.
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

	// Find links to movies pages and isolate the movie's title and year.
	// We iterate through each table row to check if it's indeed a movie
	// and not something else –– this website provides TV Series too.
	(*scraper).OnHTML("div.page-body div.wp-caption", func(e *colly.HTMLElement) {

		// Isolate the movie's title from the description
		movieName, err := utils.Normalize(e.DOM.Find("p.wp-caption-text").Text())
		if err != nil {
			log.Error.Println("Can't normalize the movie title", log.Red(err))
			return
		}

		// Isolate the movie page URL
		movieURL, urlExists := e.DOM.Find("a[href*=stills]").Attr("href")
		if !urlExists {
			log.Debug.Println("Can't find URL to movie page, next")
			return
		}

		log.Debug.Println("Found movie page link", log.White(movieURL))

		// Create folder to save images in case it doesn't exist
		moviePath, err := utils.CreateFolder(options.DataDir, options.Website, movieName)
		if err != nil {
			log.Error.Println("Can't create movie folder for:", log.White(movieName), log.Red(err))
			return
		}

		log.Info.Println("Found movie page for:", log.White(movieName))

		// Pass the movie's name and path to the next request context
		// in order to save the images in correct folder.
		ctx := colly.NewContext()
		ctx.Put("movie_name", movieName)
		ctx.Put("movie_path", moviePath)

		if err := movieScraper.Request("GET", movieURL, nil, ctx, nil); err != nil {
			log.Error.Println("Can't get movie page", log.White(movieURL), ":", log.Red(err))
		}

		// In case we enabled asynchronous jobs
		movieScraper.Wait()
	})

	// Look for links on thumbnails that redirect to a "largest" version.
	movieScraper.OnHTML("div.photo-inner dl.gallery-item a[href*=stills] img[src*=files]", func(e *colly.HTMLElement) {

		movieImageURL := e.Request.AbsoluteURL(e.Attr("src"))

		// Use regexp to remove potential GET parameters from the URL
		// regarding the resolution of the displayed image.
		movieImageURL = utils.RemoveURLParams(movieImageURL)

		log.Debug.Println("Found linked image", log.White(movieImageURL))
		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error.Println("Can't get movie image", log.White(movieImageURL), ":", log.Red(err))
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

	if err := (*scraper).Visit(StillsFrmFilmsURL); err != nil {
		log.Error.Println("Can't visit index page", log.White(StillsFrmFilmsURL), ":", log.Red(err))
	}

	// In case we enabled asynchronous jobs
	(*scraper).Wait()
}

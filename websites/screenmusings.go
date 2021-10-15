package websites

import (
	"moviestills/config"
	"moviestills/utils"
	"strings"

	log "github.com/pterm/pterm"

	"github.com/gocolly/colly/v2"
)

// Page that lists all movies available, sorted alphabetically
const ScreenMusingsURL string = "https://screenmusings.org/movie/"

// Main function that handles all the scraping logic for this website
func ScreenMusingsScraper(scraper **colly.Collector, options *config.Options) {

	// Change allowed domains for the main scrapper.
	// Images are stored on the same domain apparently.
	(*scraper).AllowedDomains = []string{
		"screenmusings.org",
	}

	// The index page might have been updated so
	// we have to revisit it when restarting scraper.
	(*scraper).AllowURLRevisit = true

	// Scraper to fetch movie images.
	// Movie pages are not updated after being published
	// therefore we only visit them once.
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

	// Isolate every movie listed, keep its title and year.
	// Create a dedicated folder if it doesn't exist to store images.
	//
	// Then visit movie page where images are listed/displayed. It seems
	// this website has both DVD and Blu-Rays reviews, let's take care of it.
	(*scraper).OnHTML("nav#movies ul li a[href*=dvd], nav#movies ul li a[href*=blu]", func(e *colly.HTMLElement) {

		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug.Println("Found movie page link", log.White(movieURL))

		// Take care of weird accents and spaces
		movieName, err := utils.Normalize(e.Text)
		if err != nil {
			log.Error.Println("Can't normalize Movie name for", log.White(e.Text), ":", log.Red(err))
			return
		}

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

		if err = movieScraper.Request("GET", movieURL, nil, ctx, nil); err != nil {
			log.Error.Println("Can't get movie page", log.White(movieURL), ":", log.Red(err))
		}

		// In case we enabled asynchronous jobs
		movieScraper.Wait()
	})

	// On every movie page, we are looking for a link to the "most viewed stills".
	// This link is extremely handy as it seems to display every thumbnail on a
	// single page. Therefore, we don't have to deal with pagination.
	movieScraper.OnHTML("ul#gallery-nav-top li:nth-last-child(2) a[href*=most]", func(e *colly.HTMLElement) {
		mostViewedImages := e.Attr("href")
		log.Debug.Println("get most viewed stills link for", log.White(e.Request.Ctx.Get("movie_name")))
		if err := e.Request.Visit(mostViewedImages); err != nil {
			log.Error.Println("Can't request most viewed stills page:", log.Red(err))
		}
	})

	// We iterate through every thumbnail on the "most viewed stills" page.
	// We have to replace "thumbnails" in the URL by "images" to get
	// the URL that links to the full resolution image.
	movieScraper.OnHTML("div#thumbnails div.thumb img[src*=thumb]", func(e *colly.HTMLElement) {

		movieImageURL := e.Request.AbsoluteURL(e.Attr("src"))

		// Replace "thumbnails" by "images" to get the full image URL
		movieImageURL = strings.Replace(movieImageURL, "thumbnails", "images", 1)

		log.Debug.Println("Found linked image", log.White(movieImageURL))
		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error.Println("Can't request linked image", log.White(movieImageURL), log.Red(err))
		}
	})

	// Check what we just visited and if its an image
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

	if err := (*scraper).Visit(ScreenMusingsURL); err != nil {
		log.Error.Println("Can't visit index page", log.White(ScreenMusingsURL), ":", log.Red(err))
	}

	// In case we enabled asynchronous jobs
	(*scraper).Wait()

}

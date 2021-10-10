package websites

import (
	"log"
	"moviestills/config"
	"moviestills/utils"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
)

// Page that lists all movies available, sorted alphabetically
const ScreenMusingsURL string = "https://screenmusings.org/movie/"

func ScreenMusingsScraper(scraper **colly.Collector, options *config.Options) {

	log.Println("Starting ScreenMusings Scraper...")

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
		log.Println(r.Request.URL, "\t", r.StatusCode, "\nError:", err)
	})

	// Before making a request print "Visiting ..."
	(*scraper).OnRequest(func(r *colly.Request) {
		log.Println("visiting index page", r.URL.String())
	})

	// Isolate every movie listed, keep its title and year.
	// Create a dedicated folder if it doesn't exist to store images.
	//
	// Then visit movie page where images are listed/displayed. It seems
	// this website has both DVD and Blu-Rays reviews, let's take care of it.
	(*scraper).OnHTML("nav#movies ul li a[href*=dvd], nav#movies ul li a[href*=blu]", func(e *colly.HTMLElement) {

		// Take care of weird accents and spaces
		movieName, err := utils.Normalize(e.Text)
		if err != nil || movieName == "" {
			log.Println("Can't normalize Movie name for", e.Text)
			return
		}

		log.Println("Found movie link for", movieName)

		// Create folder to save images in case it doesn't exist
		moviePath, err := utils.CreateFolder(options.DataDir, options.Website, movieName)
		if err != nil {
			log.Printf("Error creating folder for movie %v on %v: %v", movieName, options.Website, err)
			return
		}

		// Pass the movie's name and path to the next request context
		// in order to save the images in correct folder.
		ctx := colly.NewContext()
		ctx.Put("movie_name", movieName)
		ctx.Put("movie_path", moviePath)

		// Go to the movie page
		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("visiting movie page", movieURL)

		if err = movieScraper.Request("GET", movieURL, nil, ctx, nil); err != nil {
			log.Println("Can't visit movie page:", err)
		}

		// In case we enabled asynchronous jobs
		movieScraper.Wait()
	})

	// On every movie page, we are looking for a link to the "most viewed stills".
	// This link is extremely handy as it seems to display every thumbnail on a
	// single page. Therefore, we don't have to deal with pagination.
	movieScraper.OnHTML("ul#gallery-nav-top li:nth-last-child(2) a[href*=most]", func(e *colly.HTMLElement) {
		mostViewedImages := e.Attr("href")
		log.Println("get most viewed stills link for", e.Request.Ctx.Get("movie_name"))
		if err := e.Request.Visit(mostViewedImages); err != nil {
			log.Println("Can't request most viewed stills page:", err)
		}
	})

	// We iterate through every thumbnail on the "most viewed stills" page.
	// We have to replace "thumbnails" in the URL by "images" to get
	// the URL that links to the full resolution image.
	movieScraper.OnHTML("div#thumbnails div.thumb img[src*=thumb]", func(e *colly.HTMLElement) {

		movieImageURL := e.Request.AbsoluteURL(e.Attr("src"))

		// Replace "thumbnails" by "images" to get the full image URL
		movieImageURL = strings.Replace(movieImageURL, "thumbnails", "images", 1)

		log.Println("Found linked image", movieImageURL)
		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Println("Can't request linked image:", err)
		}
	})

	// Check what we just visited and if its an image
	// save it to the movie folder we created earlier.
	movieScraper.OnResponse(func(r *colly.Response) {

		// If we're dealing with an image, save it in the correct folder
		if strings.Contains(r.Headers.Get("Content-Type"), "image") {

			outputDir := r.Ctx.Get("movie_path")
			outputImgPath := outputDir + "/" + r.FileName()

			// Save only if we don't already downloaded it
			if _, err := os.Stat(outputImgPath); os.IsNotExist(err) {
				err = r.Save(outputImgPath)
				if err != nil {
					log.Println("Can't save image:", err)
				}
			}

			return
		}
	})

	if err := (*scraper).Visit(ScreenMusingsURL); err != nil {
		log.Println("Can't visit index page:", err)
	}

	// In case we enabled asynchronous jobs
	(*scraper).Wait()

}

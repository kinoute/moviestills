package websites

import (
	"log"
	"moviestills/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly/v2"
)

// This webpage stores a list of links to movies
const StillsFrmFilmsURL string = "https://stillsfrmfilms.wordpress.com/movies-a-z/"

func StillsFrmFilmsScraper(scraper **colly.Collector) {

	log.Println("Starting StillsFrmFilms Scraper...")

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
		log.Println(r.Request.URL, "\t", r.StatusCode, "\nError:", err)
	})

	// Before making a request print "Visiting ..."
	(*scraper).OnRequest(func(r *colly.Request) {
		log.Println("visiting index page", r.URL.String())
	})

	// Find links to movies pages and isolate the movie's title and year.
	// We iterate through each table row to check if it's indeed a movie
	// and not something else –– this website provides TV Series too.
	(*scraper).OnHTML("div.page-body div.wp-caption", func(e *colly.HTMLElement) {

		// Isolate the movie's title from the description
		movieName, _ := utils.Normalize(e.DOM.Find("p.wp-caption-text").Text())

		// Isolate the movie page URL
		movieURL, _ := e.DOM.Find("a[href*=stills]").Attr("href")

		log.Println("Found movie page link", movieURL)

		// Create folder to save images in case it doesn't exist
		moviePath := filepath.Join(".", "data", "stillsfrmfilms", movieName)
		err := os.MkdirAll(moviePath, os.ModePerm)
		if err != nil {
			log.Println("Error creating folder for", movieName)
			return
		}

		// Pass the movie's name and path to the next request context
		// in order to save the images in correct folder.
		ctx := colly.NewContext()
		ctx.Put("movie_name", movieName)
		ctx.Put("movie_path", moviePath)

		movieScraper.Request("GET", movieURL, nil, ctx, nil)

		movieScraper.Visit(movieURL)
	})

	// Look for links on thumbnails that redirect to a "largest" version.
	movieScraper.OnHTML("div.photo-inner dl.gallery-item a[href*=stills]", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found linked image", movieImageURL)
		e.Request.Visit(movieImageURL)
	})

	// Check if what we just visited is an image and
	// save it to the movie folder we created earlier.
	movieScraper.OnResponse(func(r *colly.Response) {

		if strings.Index(r.Headers.Get("Content-Type"), "image") > -1 {
			outputDir := r.Ctx.Get("movie_path")
			outputImgPath := outputDir + "/" + r.FileName()

			// Don't save again it we already downloaded it
			if _, err := os.Stat(outputImgPath); os.IsNotExist(err) {
				r.Save(outputImgPath)
			}
			return
		}
	})

	(*scraper).Visit(StillsFrmFilmsURL)
}
package websites

import (
	"log"
	"moviestills/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly/v2"
)

// This webpage stores a list of links to movie, TV movies, Series...
const EvanERichardsURL string = "https://www.evanerichards.com/index"

func EvanERichardsScraper(scraper **colly.Collector) {

	log.Println("Starting Evan E. Richards Scraper...")

	// Change allowed domain for the main scraper.
	// Since everything is served on the same domain,
	// only one domain is really necessary.
	(*scraper).AllowedDomains = []string{
		"www.evanerichards.com",
		"evanerichards.com",
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
	(*scraper).OnHTML("tbody tr.pp-table-row", func(e *colly.HTMLElement) {

		// Fetch various datas in columns for each table entry
		title, _ := utils.Normalize(e.DOM.Find("td.pp-table-cell-Title a").Text())
		category, _ := utils.Normalize(e.DOM.Find("td.pp-table-cell-Category").Text())
		year, _ := utils.Normalize(e.DOM.Find("td.pp-table-cell-Date").Text())

		// Ignore entries that are not movies
		if category != "Movie" && category != "Animation" {
			log.Printf("\"%s\" is not a Movie, ignoring...", title)
			return
		}

		movieURL, _ := e.DOM.Find("td.pp-table-cell-Title a").Attr("href")

		log.Println("Found movie page link", movieURL)

		// Create folder to save images in case it doesn't exist
		moviePath := filepath.Join(".", "data", "evanerichards", title)
		if err := os.MkdirAll(moviePath, os.ModePerm); err != nil {
			log.Println("Error creating folder for", title)
			return
		}

		// Pass the movie's name, year and path to the next request context
		// in order to save the images in correct folder.
		ctx := colly.NewContext()
		ctx.Put("movie_name", title)
		ctx.Put("movie_year", year)
		ctx.Put("movie_path", moviePath)

		if err := movieScraper.Request("GET", movieURL, nil, ctx, nil); err != nil {
			log.Println("Can't visit movie page:", err)
		}

		// In case we enabled asynchronous jobs
		movieScraper.Wait()
	})

	// Look for links on thumbnails that redirect to a "largest" version.
	movieScraper.OnHTML("div.elementor-widget-container div.ngg-gallery-thumbnail a[class*=shutter]", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found linked image", movieImageURL)

		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Println("Can't request linked image:", err)
		}
	})

	// Check if what we just visited is an image and
	// save it to the movie folder we created earlier.
	movieScraper.OnResponse(func(r *colly.Response) {
		if strings.Contains(r.Headers.Get("Content-Type"), "image") {

			outputDir := r.Ctx.Get("movie_path")
			outputImgPath := outputDir + "/" + r.FileName()

			// Don't save again it we already downloaded it
			if _, err := os.Stat(outputImgPath); os.IsNotExist(err) {
				if err = r.Save(outputImgPath); err != nil {
					log.Println("Can't save image:", err)
				}
			}

			return
		}
	})

	if err := (*scraper).Visit(EvanERichardsURL); err != nil {
		log.Println("Can't visit index page:", err)
	}

	// In case we enabled asynchronous jobs
	(*scraper).Wait()
}

package websites

import (
	"moviestills/config"
	"moviestills/utils"
	"strings"

	log "github.com/pterm/pterm"

	"github.com/gocolly/colly/v2"
)

// EvanERichardsURL is the webpage that stores a list of links to movie,
// TV movies, Series...
const EvanERichardsURL string = "https://www.evanerichards.com/index"

// EvanERichardsScraper is the main function that handles all the scraping logic
// for this website.
func EvanERichardsScraper(scraper **colly.Collector, options *config.Options) {

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

	if err := (*scraper).Visit(EvanERichardsURL); err != nil {
		log.Error.Println("Can't visit index page", log.White(BluBeaverURL), ":", log.Red(err))
	}

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
	(*scraper).OnHTML("tbody tr.pp-table-row", func(e *colly.HTMLElement) {

		// Fetch various datas in columns for each table entry
		title, _ := utils.Normalize(e.DOM.Find("td.pp-table-cell-Title a").Text())
		category, _ := utils.Normalize(e.DOM.Find("td.pp-table-cell-Category").Text())
		year, _ := utils.Normalize(e.DOM.Find("td.pp-table-cell-Date").Text())

		// Ignore entries that are not movies
		if category != "Movie" && category != "Animation" {
			log.Debug.Println(log.White(title), "is not a Movie, ignoring...")
			return
		}

		movieURL, urlExists := e.DOM.Find("td.pp-table-cell-Title a").Attr("href")
		if !urlExists {
			log.Debug.Println("could not find movie URL, next")
			return
		}

		log.Debug.Println("Found movie page link", log.White(movieURL))

		// Create folder to save images in case it doesn't exist
		moviePath, err := utils.CreateFolder(options.DataDir, options.Website, title)
		if err != nil {
			log.Error.Println("Can't create movie folder for:", log.White(title), log.Red(err))
			return
		}

		log.Info.Println("Found movie page for:", log.White(title))

		// Pass the movie's name, year and path to the next request context
		// in order to save the images in correct folder.
		ctx := colly.NewContext()
		ctx.Put("movie_name", title)
		ctx.Put("movie_year", year)
		ctx.Put("movie_path", moviePath)

		if err := movieScraper.Request("GET", movieURL, nil, ctx, nil); err != nil {
			log.Error.Println("Can't get movie page", log.White(movieURL), ":", log.Red(err))
		}

	})

	// Look for links on thumbnails that redirect to a "largest" version
	movieScraper.OnHTML("div.elementor-widget-container div.ngg-gallery-thumbnail a[class*=shutter]", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug.Println("Found linked image", log.White(movieImageURL))

		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error.Println("Can't get large image", log.White(movieImageURL), ":", log.Red(err))
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

	// Ensure that all requests are completed before exiting
	if (*scraper).Async {
		(*scraper).Wait()
		movieScraper.Wait()
	}

}

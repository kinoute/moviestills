package websites

import (
	"log"
	"moviestills/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

// Page that lists all movies available, sorted alphabetically
const ScreenCapsURL string = "https://movie-screencaps.com/movie-directory/"

func ScreenCapsScraper(scraper **colly.Collector) {

	log.Println("Starting Movie-Screencaps Scraper...")

	// Change allowed domains for the main scrapper.
	// Images are stored on wordpress apparently.
	(*scraper).AllowedDomains = []string{
		"movie-screencaps.com",
		"www.movie-screencaps.com",
		"i0.wp.com",
		"i1.wp.com",
		"i2.wp.com",
		"wp.com",
	}

	// The index page might have been updated so
	// we have to revisit it when restarting scraper.
	(*scraper).AllowURLRevisit = true

	// Scraper to fetch movie images.
	// Movie pages are not updated after being published therefore we only visit once.
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

	// Isolate every movie listed, keep its title and
	// create a dedicated folder if it doesn't exist
	// to store images.
	//
	// Then visit movie page where images are listed/displayed.
	(*scraper).OnHTML("div.tagindex ul.links li a[href*=movie]", func(e *colly.HTMLElement) {

		movieName, err := utils.Normalize(e.Text)
		if err != nil || movieName == "" {
			log.Println("Can't normalize Movie name for", e.Text)
			return
		}

		log.Println("Found movie link for", movieName)

		// Create folder to save images in case it doesn't exist
		moviePath := filepath.Join(".", "data", "movie-screencaps", movieName)
		err = os.MkdirAll(moviePath, os.ModePerm)
		if err != nil {
			log.Println("Error creating folder for", movieName)
			return
		}

		// Pass the movie's name and path to the next request context
		// in order to save the images in correct folder.
		ctx := colly.NewContext()
		ctx.Put("movie_name", movieName)
		ctx.Put("movie_path", moviePath)

		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("visiting movie page", movieURL)

		movieScraper.Request("GET", movieURL, nil, ctx, nil)
	})

	// Handle pagination by getting the number of pages in total first.
	// Then iterate through all pages with a for loop to get movie stills.
	movieScraper.OnHTML("div.pixcode + div.wp-pagenavi > select.paginate option:last-child", func(e *colly.HTMLElement) {

		// Get the URL of the movie page
		actualPageURL := e.Request.URL.String()

		// Only start to visit paginated pages if we're at the first page.
		// Otherwise it will result in an infinite loop.
		if strings.Index(actualPageURL, "/page") == -1 {

			// Get the total number of pages from the select menu and the last option
			numOfPages, _ := strconv.Atoi(e.Attr("value"))
			log.Println("number of pages for", e.Request.Ctx.Get("movie_name"), "is", e.Attr("value"))

			// Visit every paginated page to get a few snapshots every time
			for num := 2; num <= numOfPages; num++ {
				log.Println("visiting paginated page", strconv.Itoa(num), "for", e.Request.Ctx.Get("movie_name"))
				e.Request.Visit(actualPageURL + "page/" + strconv.Itoa(num))
			}
		}

	})

	// Go through each link to a movie snapshot found on the movie page.
	//
	// Exceptionnally, since this website basically takes a snapshot every
	// second or so during the movie, if we download everything, we will have
	// many similar snapshots and it's going to take forever.
	//
	// Therefore, we added :nth-of-type(40n) to the CSS selector to only
	// download 1 shot out of 40. Remove it if you want to download everything.
	movieScraper.OnHTML("section.entry-content a[href*=wp][href*=caps]:nth-of-type(40n)", func(e *colly.HTMLElement) {

		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("inside movie page for", e.Request.Ctx.Get("movie_name"))

		log.Println("Found linked image", movieImageURL)
		e.Request.Visit(movieImageURL)
	})

	// Check what we just visited and if its an image
	// save it to the movie folder we created earlier.
	movieScraper.OnResponse(func(r *colly.Response) {

		// If we're dealing with an image, save it in the correct folder
		if strings.Index(r.Headers.Get("Content-Type"), "image") > -1 {

			outputDir := r.Ctx.Get("movie_path")

			// We're getting weird filenames from Wordpress.
			// We might need to remove some suffix.
			fileName := strings.TrimSuffix(r.FileName(), "_strip_all")
			outputImgPath := outputDir + "/" + fileName

			// Save only if we don't already downloaded it
			if _, err := os.Stat(outputImgPath); os.IsNotExist(err) {
				r.Save(outputImgPath)
			}

			return
		}
	})

	(*scraper).Visit(ScreenCapsURL)

}

package websites

import (
	"log"
	"moviestills/config"
	"moviestills/utils"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

// Page that lists all movies available, sorted alphabetically
const ScreenCapsURL string = "https://movie-screencaps.com/movie-directory/"

func ScreenCapsScraper(scraper **colly.Collector, options *config.Options) {

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

		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("visiting movie page", movieURL)

		if err = movieScraper.Request("GET", movieURL, nil, ctx, nil); err != nil {
			log.Println("Can't visit movie page:", err)
		}

		// In case we enabled asynchronous jobs
		movieScraper.Wait()
	})

	// Handle pagination by getting the number of pages in total first.
	// Then iterate through all pages with a for loop to get movie stills.
	movieScraper.OnHTML("div.pixcode + div.wp-pagenavi > select.paginate option:last-child", func(e *colly.HTMLElement) {

		// Get the URL of the movie page
		actualPageURL := e.Request.URL.String()

		// Only start to visit paginated pages if we're at the first page.
		// Otherwise it will result in an infinite loop.
		if !strings.Contains(actualPageURL, "/page") {

			// Get the total number of pages from the select menu and the last option
			numOfPages, _ := strconv.Atoi(e.Attr("value"))
			log.Println("number of pages for", e.Request.Ctx.Get("movie_name"), "is", e.Attr("value"))

			// Visit every paginated page to get a few snapshots every time
			for num := 2; num <= numOfPages; num++ {
				log.Println("visiting paginated page", strconv.Itoa(num), "for", e.Request.Ctx.Get("movie_name"))
				if err := e.Request.Visit(actualPageURL + "page/" + strconv.Itoa(num)); err != nil {
					log.Println("Can't visit paginated page:", err)
				}
			}
		}

	})

	// Go through each link to a movie snapshot found on the movie page.
	//
	// Exceptionnally, since this website basically takes a snapshot every
	// second or so during the movie, if we download everything, we will have
	// many similar snapshots and it's going to take forever.
	//
	// Therefore, we added :nth-of-type(30n) to the CSS selector to only
	// download 1 shot every 30 shots. Remove it if you want to download everything.
	movieScraper.OnHTML("section.entry-content a[href*=wp][href*=caps]:nth-of-type(30n)", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))

		// We're getting weird filenames from Wordpress with "strip=all" at the end.
		// We might need to remove some suffixes.
		movieImageURL = utils.RemoveURLParams(movieImageURL)

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
				if err = r.Save(outputImgPath); err != nil {
					log.Println("Can't save image:", err)
				}
			}

			return
		}
	})

	if err := (*scraper).Visit(ScreenCapsURL); err != nil {
		log.Println("Can't visit index page:", err)
	}

	// In case we enabled asynchronous jobs
	(*scraper).Wait()

}

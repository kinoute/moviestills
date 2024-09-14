package websites

import (
	"moviestills/config"
	"moviestills/utils"
	"strconv"
	"strings"

	log "github.com/pterm/pterm"

	"github.com/gocolly/colly/v2"
)

// ScreenCapsURL is the page that lists all movies available, sorted alphabetically
const ScreenCapsURL string = "https://movie-screencaps.com/movie-directory/"

// ScreenCapsScraper is the main function that handles all the scraping logic
// for this website.
func ScreenCapsScraper(scraper **colly.Collector, options *config.Options) {

	// Change allowed domains for the main scrapper.
	// Images are stored on wordpress apparently.
	(*scraper).AllowedDomains = []string{
		"movie-screencaps.com",
		"www.movie-screencaps.com",
		"i0.wp.com",
		"i1.wp.com",
		"i2.wp.com",
		"i3.wp.com",
		"wp.com",
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

	// Isolate every movie listed, keep its title and
	// create a dedicated folder if it doesn't exist
	// to store images.
	//
	// Then visit movie page where images are listed/displayed.
	(*scraper).OnHTML("div.tagindex ul.links li a[href*=movie]", func(e *colly.HTMLElement) {

		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug.Println("Found movie page link", log.White(movieURL))

		// Take care of weird accents and spaces
		movieName, err := utils.Normalize(e.Text)
		if err != nil {
			log.Error.Println("Can't normalize Movie name for", log.White(e.Text), ":", log.Red(err))
			return
		}

		log.Debug.Println("Found movie link for", log.White(movieName))

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

	})

	// Handle pagination by getting the number of pages in total first.
	// Then iterate through all pages with a for loop to get movie stills.
	movieScraper.OnHTML("div.pixcode + div.wp-pagenavi > select.paginate option:last-child", func(e *colly.HTMLElement) {

		// Get the URL of the movie page
		actualPageURL := e.Request.URL.String()

		// Only start to visit paginated pages if we're at the first page.
		// Otherwise it will result in an infinite loop.
		if !strings.Contains(actualPageURL, "/page") {

			movieName := e.Request.Ctx.Get("movie_name")

			// Get the total number of pages from the select menu and the last option
			numOfPages, _ := strconv.Atoi(e.Attr("value"))
			log.Info.Println("number of pages for", log.White(movieName), "is", log.White(e.Attr("value")))

			// Visit every paginated page to get a few snapshots every time
			for num := 2; num <= numOfPages; num++ {
				log.Info.Println("visiting paginated page", log.White(strconv.Itoa(num)), "for", log.White(movieName))
				paginatedPageURL := actualPageURL + "page/" + strconv.Itoa(num)
				if err := e.Request.Visit(paginatedPageURL); err != nil {
					log.Error.Println("Can't visit paginated page", log.White(paginatedPageURL), ":", log.Red(err))
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

		log.Debug.Println("Found linked image", log.White(movieImageURL))
		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error.Println("Can't request linked image", log.White(movieImageURL), log.Red(err))
		}
	})

	// Before making a request to URL
	movieScraper.OnRequest(func(r *colly.Request) {
		log.Debug.Println("visiting", log.White(r.URL.String()))
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

	if err := (*scraper).Visit(ScreenCapsURL); err != nil {
		log.Error.Println("Can't visit index page", log.White(ScreenCapsURL), ":", log.Red(err))
	}

	// Ensure that all requests are completed before exiting
	if (*scraper).Async {
		(*scraper).Wait()
		movieScraper.Wait()
	}

}

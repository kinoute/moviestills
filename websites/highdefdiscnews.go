package websites

import (
	"moviestills/config"
	"moviestills/utils"
	"strings"

	log "github.com/pterm/pterm"

	"github.com/gocolly/colly/v2"
)

// This webpage stores a list of links to movie pages with Blu-rays images
const HighDefDiscNewsURL string = "https://highdefdiscnews.com/blu-ray-screenshots/"

// Main function that handles all the scraping logic for this website
func HighDefDiscNewsScraper(scraper **colly.Collector, options *config.Options) {

	// Change allowed domain for the main scraper.
	// Since everything is served on the same domain,
	// only one domain is really necessary.
	(*scraper).AllowedDomains = []string{
		"highdefdiscnews.com",
	}

	// The index page might have been updated since last visit so
	// we have to revisit it when restarting the scraper.
	// It is a single page, it will not cost anything anyway.
	(*scraper).AllowURLRevisit = true

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

	// Find links to movies reviews and isolate the movie's title.
	// Links contain some useless text such as "- Blu-ray Screenshots"
	// or "[Remastered]".
	// We remove these texts to isolate the movie's title.
	(*scraper).OnHTML("div#mcTagMap ul.links a[href*=high]", func(e *colly.HTMLElement) {
		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug.Println("Found movie page link", log.White(movieURL))

		// Isolate the movie's title from the text
		tmpMovieName := isolateMovieTitle(e.Text)

		// Remove weird accents and spaces from the movie's title
		movieName, err := utils.Normalize(tmpMovieName)
		if err != nil {
			log.Error.Println("Can't normalize Movie name for", log.White(e.Text), log.Red(err))
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

		if err := movieScraper.Request("GET", movieURL, nil, ctx, nil); err != nil {
			log.Error.Println("Can't visit movie page", log.White(movieURL), ":", log.Red(err))
		}

		// In case we enabled asynchronous jobs
		movieScraper.Wait()
	})

	// Look for links on thumbnails that redirects to a "largest" version.
	movieScraper.OnHTML("div.gallery dl.gallery-item a[href*=high]", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug.Println("Found linked image", log.White(movieImageURL))
		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error.Println("Can't get linked image", log.White(movieImageURL), ":", log.Red(err))
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

	if err := (*scraper).Visit(HighDefDiscNewsURL); err != nil {
		log.Error.Println("Can't visit index page", log.White(HighDefDiscNewsURL), ":", log.Red(err))
	}

	// In case we enabled asynchronous jobs
	(*scraper).Wait()
}

// Isolate the movie's title by getting rid of various words on the right.
// eg. "[Remastered]", "- Blu-ray Screenshots".
func isolateMovieTitle(sentence string) string {

	// Get rid of text such as "- Blu-ray Screenshots"
	if idx := strings.LastIndex(sentence, " - "); idx != -1 {
		sentence = sentence[:idx]
	}

	// Get rid of text such as "[Remastered]"
	if idx := strings.LastIndex(sentence, " ["); idx != -1 {
		sentence = sentence[:idx]
	}

	return sentence
}

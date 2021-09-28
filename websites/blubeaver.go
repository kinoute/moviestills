package websites

import (
	"log"
	"moviestills/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly/v2"
)

// This webpage stores a list of links to movie reviews of Blu-rays
const BluBeaverURL string = "http://www.dvdbeaver.com/blu-ray.htm"

func BluBeaverScraper(scraper **colly.Collector) {

	log.Println("Starting BluBeaver Scraper...")

	// Change allowed domain for the main scraper.
	// Since everything is served on the same domain,
	// only one domain is really necessary.
	(*scraper).AllowedDomains = []string{
		"www.blubeaver.ca",
		"www.dvdbeaver.com",
		"dvdbeaver.com",
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
		log.Println(r.Request.URL, "\t", r.StatusCode, "\nError:", err)
	})

	// Before making a request print "Visiting ..."
	(*scraper).OnRequest(func(r *colly.Request) {
		log.Println("visiting index page", r.URL.String())
	})

	// Find links to movies reviews and isolate the movie's title.
	// Since DVDBeaver is somewhat a custom website, some links
	// might have different cases. We use the CSS4 "i" case-insensitive
	// feature to make sure our filter doesn't miss anything.
	(*scraper).OnHTML("tr b a[href*='film' i][href*='review' i]", func(e *colly.HTMLElement) {
		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found movie page link", movieURL)

		// Remove weird accents and spaces from the movie's title
		movieName, err := utils.Normalize(e.Text)
		if err != nil || movieName == "" {
			log.Println("Can't normalize Movie name for", e.Text)
			return
		}

		// Create folder to save images in case it doesn't exist
		moviePath := filepath.Join(".", "data", "blubeaver", movieName)
		if err = os.MkdirAll(moviePath, os.ModePerm); err != nil {
			log.Println("Error creating folder for", movieName)
			return
		}

		// Pass the movie's name and path to the next request context
		// in order to save the images in correct folder.
		ctx := colly.NewContext()
		ctx.Put("movie_name", movieName)
		ctx.Put("movie_path", moviePath)

		if err = movieScraper.Request("GET", movieURL, nil, ctx, nil); err != nil {
			log.Println("Can't request movie page", err)
		}

		// In case we enabled asynchronous jobs
		movieScraper.Wait()
	})

	// Look for links on images that redirects to a "largest" version.
	// These links appear on Blu-Ray reviews almost exclusively and
	// provide images with native resolution (1080p).
	//
	// We try to avoid images with "subs" in the filename as they are
	// most likely images with subtitles on top. We don't want that.
	movieScraper.OnHTML("a[href*='large' i]:not([href*='subs' i])", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found linked image", movieImageURL)
		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Println("Can't request linked image", err)
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
					log.Println("Can't save image", err)
				}
			}
			return
		}
	})

	if err := (*scraper).Visit(BluBeaverURL); err != nil {
		log.Println("Can't visit index page", err)
	}

	// In case we enabled asynchronous jobs
	(*scraper).Wait()
}

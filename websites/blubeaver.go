package websites

import (
	"log"
	"moviestills/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly/v2"
)

// this webpage stores a list of links to
// movie reviews of Blu-rays
const BluBeaverURL string = "http://www.dvdbeaver.com/blu-ray.htm"

func BluBeaverScraper(scraper **colly.Collector) {
	log.Println("Starting BluBeaver Scraper...")

	// change allowed domain for the main scrapper
	// since everything is served on the same domain,
	// only one domain is necessary.
	(*scraper).AllowedDomains = []string{"www.blubeaver.ca", "www.dvdbeaver.com", "dvdbeaver.com"}

	// the index page might have been changed so
	// we have to revisit it when restarting the scraper.
	// it's a single page, it will not cost anything anyway.
	(*scraper).AllowURLRevisit = true

	// scraper to fetch movie images on reviews pages.
	// These pages are not updated after being
	// published therefore we only visit them once
	movieScraper := (*scraper).Clone()
	movieScraper.AllowURLRevisit = false

	(*scraper).OnError(func(r *colly.Response, err error) {
		log.Println(r.Request.URL, "\t", r.StatusCode, "\nError:", err)
	})

	// Before making a request print "Visiting ..."
	(*scraper).OnRequest(func(r *colly.Request) {
		log.Println("visiting index page", r.URL.String())
	})

	// find links to movies reviews
	(*scraper).OnHTML("a[href*='film' i][href*='review' i]", func(e *colly.HTMLElement) {
		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found movie page link", movieURL)

		movieName, err := utils.Normalize(e.Text)
		if err != nil || movieName == "" {
			log.Println("Can't normalize Movie name for", movieName)
			return
		}

		// create folder to save images in case it doesn't exist
		moviePath := filepath.Join(".", "data", "blubeaver", movieName)
		err = os.MkdirAll(moviePath, os.ModePerm)
		if err != nil {
			log.Println("Error creating folder for", movieName)
			return
		}

		// pass the movie's name and path to the next request context
		// in order to save the images in correct folder
		ctx := colly.NewContext()
		ctx.Put("movie_name", movieName)
		ctx.Put("movie_path", moviePath)

		movieScraper.Request("GET", movieURL, nil, ctx, nil)

		movieScraper.Visit(movieURL)
	})

	// look for links on images that redirects to a "largest" version.
	// most likely, these links appear on Blu-Ray reviews.
	movieScraper.OnHTML("a[href*='large' i]:not([href*='subs' i])", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found linked image", movieImageURL)
		e.Request.Visit(movieImageURL)
	})

	// check what we just visited and if its an image
	// save it to the movie folder we created earlier
	movieScraper.OnResponse(func(r *colly.Response) {
		if strings.Index(r.Headers.Get("Content-Type"), "image") > -1 {
			outputDir := r.Ctx.Get("movie_path")
			outputImgPath := outputDir + "/" + r.FileName()

			// don't save again it we already downloaded it
			if _, err := os.Stat(outputImgPath); os.IsNotExist(err) {
				r.Save(outputImgPath)
			}
			return
		}
	})

	(*scraper).Visit(BluBeaverURL)
}

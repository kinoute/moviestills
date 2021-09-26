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

// this webpage stores a list of links to
// movie list pages by alphabet (#, a, z).
// It's a good starting point for our task.
const BeaverURL string = "http://www.dvdbeaver.com/film/reviews.htm"

func DVDBeaverScraper(scraper **colly.Collector) {
	log.Println("Starting DVDBeaver Scraper...")

	// change allowed domain for the main scrapper
	// since everything is served on the same domain,
	// only one domain is necessary.
	(*scraper).AllowedDomains = []string{"www.dvdbeaver.com"}

	// the reviews page might have been changed so
	// we have to revisit it when restarting the scraper.
	// it's a single page, it will not cost anything anyway.
	(*scraper).AllowURLRevisit = true

	// movie list might be updated often with new movies
	// so we autorize the scraper to revisit these pages.
	movieListScraper := (*scraper).Clone()
	movieListScraper.AllowURLRevisit = true

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
		log.Println("visiting reviews page", r.URL.String())
	})

	// find links to movies list by alphabet
	(*scraper).OnHTML("a[href*='listing' i]", func(e *colly.HTMLElement) {
		movieListURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found movie list page link", movieListURL)
		movieListScraper.Visit(movieListURL)
	})

	// Before making a request print "Visiting ..."
	movieListScraper.OnRequest(func(r *colly.Request) {
		log.Println("visiting movie list page", r.URL.String())
	})

	// looks for movie reviews link and create folder for every
	// movie we find to prepare the download of the snapshots.
	movieListScraper.OnHTML("a[href*='film' i][href*='review' i]", func(e *colly.HTMLElement) {

		// take care of weird characters
		movieName, err := utils.Normalize(e.Text)
		if err != nil || movieName == "" {
			log.Println("Can't normalize Movie name for", e.Text)
			return
		}

		log.Println("Found movie link for ", movieName)

		// create folder to save images in case it doesn't exist yet
		moviePath := filepath.Join(".", "data", "dvdbeaver", movieName)
		err = os.MkdirAll(moviePath, os.ModePerm)
		if err != nil {
			log.Println("Error creating folder for", movieName)
			return
		}

		// pass the movie path to the next request context
		// in order to save the images in the correct folder
		ctx := colly.NewContext()
		ctx.Put("movie_path", moviePath)

		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("visiting movie page", movieURL)

		movieScraper.Request("GET", movieURL, nil, ctx, nil)
	})

	// look for links on images that redirects to a "largest" version.
	// most likely, these links appear on Blu-Ray reviews.
	movieScraper.OnHTML("a[href*='large' i]:not([href*='subs' i])", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found linked image", movieImageURL)
		e.Request.Visit(movieImageURL)
	})

	// on DVD reviews, there is no clickable large version
	// so we download the images as shown on the webpage and
	// be sure we avoid some weird images (subtitles, covers etc)
	movieScraper.OnHTML(
		"img:not([src*='banner' i])"+
			":not([src*='bitrate' i])"+
			":not([src$='gif' i])"+
			":not([src*='subs' i])"+
			":not([src*='daggers' i])"+
			":not([src*='posters' i])"+
			":not([src*='menu' i])", func(e *colly.HTMLElement) {
			movieImageURL := e.Request.AbsoluteURL(e.Attr("src"))

			// filter low resolutions images to avoid false positives
			// if the images are too small, we won't be able to use them
			// anyway so lets skip them.
			movieImageWidth, _ := strconv.Atoi(e.Attr("width"))
			movieImageHeight, _ := strconv.Atoi(e.Attr("height"))
			if movieImageHeight >= 275 && movieImageWidth >= 500 {
				log.Println("Image seems correct in sizes, downloading", movieImageURL)
				e.Request.Visit(movieImageURL)
			}
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

	(*scraper).Visit(BeaverURL)
}

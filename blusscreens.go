package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly/v2"
)

// we use the cinematographers page because the movie's title
// can be easily scrapped/parsed there instead of the screen
// captures page (#, A, Z) where links are images.
// The downside is, some movies (animation?) might be missing
// because they don't have cinematographers associated to it.
var BlusURL = "https://www.bluscreens.net/cinematographers.html"

func newBlusScraper(scraper **colly.Collector) {

	log.Println("Starting Blus Screens Scraper...")

	// change allowed domains for the main scrapper
	// images are stored on imgur so make sure we allow it
	(*scraper).AllowedDomains = []string{"www.bluscreens.net", "imgur.com", "i.imgur.com"}

	// the cinematographers page might have been changed so
	// we have to revisit it when restarting scraper
	(*scraper).AllowURLRevisit = true

	// scraper to fetch movie images
	// movie page is not updated after being
	// published therefore we only visit once
	movieScraper := (*scraper).Clone()
	movieScraper.AllowURLRevisit = false

	// Before making a request print "Visiting ..."
	(*scraper).OnRequest(func(r *colly.Request) {
		log.Println("visiting movie list page", r.URL.String())
	})

	// isolate every movie listed, keep its title and
	// create a dedicated folder if it doesn't exist
	// to store images. Then visit movie page where
	// images are listed/displayed
	(*scraper).OnHTML("h2.wsite-content-title a[href*=html]", func(e *colly.HTMLElement) {

		movieName := strings.TrimSpace(e.Text)

		log.Println("Found movie link for", movieName)

		// create folder to save images in case it doesn't exist
		moviePath := filepath.Join(".", "data", "blusscreens", movieName)
		err := os.MkdirAll(moviePath, os.ModePerm)
		if err != nil {
			log.Println("Error creating folder for", movieName)
			return
		}

		// pass the movie's name and path to the next request context
		// in order to save the images in correct folder
		ctx := colly.NewContext()
		ctx.Put("movie_path", moviePath)

		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("visiting movie page", movieURL)

		movieScraper.Request("GET", movieURL, nil, ctx, nil)
	})

	// go through each link to imgur found on the movie page
	movieScraper.OnHTML("div.galleryInnerImageHolder a[href*=imgur]", func(e *colly.HTMLElement) {

		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("inside movie page for", e.Request.Ctx.Get("movie_path"))

		// create link to real image if its a link to imgur
		// website and not directly to the image
		// eg. https://imgur.com/ABC to https://i.imgur.com/ABC.png
		if strings.Index(movieImageURL, "i.imgur.com") < 0 {
			movieImageURL += ".png"
			movieImageURL = strings.Replace(movieImageURL, "https://imgur.com", "https://i.imgur.com", 1)
		}

		log.Println("Found linked image", movieImageURL)
		e.Request.Visit(movieImageURL)
	})

	// check what we just visited and if its an image
	// save it to the movie folder we created earlier
	movieScraper.OnResponse(func(r *colly.Response) {

		// if we're dealing with an image, save it in the correct folder
		if strings.Index(r.Headers.Get("Content-Type"), "image") > -1 {
			outputDir := r.Ctx.Get("movie_path")
			outputImgPath := outputDir + "/" + r.FileName()

			// save only if we don't already downloaded it
			if _, err := os.Stat(outputImgPath); os.IsNotExist(err) {
				r.Save(outputImgPath)
			}
			return
		}
	})

	(*scraper).Visit(BlusURL)
}

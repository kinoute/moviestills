package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

var url = "http://www.dvdbeaver.com/film/reviews.htm"

func newBeaverScraper(scraper **colly.Collector) {
	log.Println("inside beaver scraper")

	// change allowed domain for the main scrapper
	(*scraper).AllowedDomains = []string{"www.dvdbeaver.com"}
	// the reviews page might have been changed so
	// we have to revisit it when restarting scraper
	(*scraper).AllowURLRevisit = true

	// movie list might be updated often with new movies
	// so we autorize scraper to revisit
	movieListScraper := (*scraper).Clone()
	movieListScraper.AllowURLRevisit = true

	// scraper to fetch movie images
	// this type of page is not updated after being
	// published therefore we only visit once
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
	(*scraper).OnHTML("a[href*=listing]", func(e *colly.HTMLElement) {
		movieListURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found movie list page link", movieListURL)
		movieListScraper.Visit(movieListURL)
	})

	// Before making a request print "Visiting ..."
	movieListScraper.OnRequest(func(r *colly.Request) {
		log.Println("visiting movie list page", r.URL.String())
	})

	movieListScraper.OnHTML("a[href*=film][href*=review]", func(e *colly.HTMLElement) {
		movieName := strings.TrimSpace(e.Text)

		log.Println("Found movie link for ", movieName)

		// create folder to save images in case it doesn't exist
		moviePath := filepath.Join(".", "data", "dvdbeaver", movieName)
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

	// look for images linked to a "large" version
	movieScraper.OnHTML("a[href*=large]:not([href*=subs])", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found linked image", movieImageURL)
		e.Request.Visit(movieImageURL)
	})

	// on DVD reviews, there is no clickable large version
	// download the image as shown on the webpage
	movieScraper.OnHTML("td img:not([src*=banner]):not([src*=bitrate]):not([src$=gif]):not([src*=subs]):not([src*=daggers]):not([src*=menu])", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("src"))

		// filter low resolutions images to avoid false positive
		movieImageWidth, _ := strconv.Atoi(e.Attr("width"))
		movieImageHeight, _ := strconv.Atoi(e.Attr("height"))

		log.Println("Found low image", movieImageURL)

		if movieImageHeight >= 275 && movieImageWidth >= 500 {
			log.Println("Image seems correct in sizes, downloading")
			e.Request.Visit(movieImageURL)
		}
	})

	movieScraper.OnResponse(func(r *colly.Response) {
		if strings.Index(r.Headers.Get("Content-Type"), "image") > -1 {
			outputDir := r.Ctx.Get("movie_path")
			r.Save(outputDir + "/" + r.FileName())
			return
		}
	})

	(*scraper).Visit(url)

	// if visited, _ := (*scraper).HasVisited(url); !visited {
	// 	log.Println("not visited, lets go", url)
	// 	(*scraper).Visit(url)
	// } else {
	// 	log.Println("already visited", url)
	// }

}

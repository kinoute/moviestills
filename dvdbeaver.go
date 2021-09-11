package main

import (
	"log"
	"strings"

	"github.com/gocolly/colly/v2"
)

var url = "http://www.dvdbeaver.com/film/reviews.htm"

func newBeaverScraper(scraper **colly.Collector) {
	log.Println("inside beaver scraper")

	// change allowed domain from the main scrapper
	(*scraper).AllowedDomains = []string{"www.dvdbeaver.com"}
	// the reviews page might have been changed so
	// we have to revisti it
	(*scraper).AllowURLRevisit = true

	// movie list might be updated often with new movies
	// autorize revisit
	movieListScraper := (*scraper).Clone()
	movieListScraper.AllowURLRevisit = true

	// scraper to fetch movie images
	// this type of page is not updated after being
	// published therefore we only visit once
	movieScraper := (*scraper).Clone()

	(*scraper).OnResponse(func(r *colly.Response) {
		log.Println(r.Request.URL, "\t", r.StatusCode)
	})

	(*scraper).OnError(func(r *colly.Response, err error) {
		log.Println(r.Request.URL, "\t", r.StatusCode, "\nError:", err)
	})

	(*scraper).OnScraped(func(r *colly.Response) {
		log.Println("Finished", r.Request.URL)
	})

	// find links to movies list by alphabet
	(*scraper).OnHTML("a[href*=listing]", func(e *colly.HTMLElement) {
		movieListURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found movie list page link", movieListURL)
		movieListScraper.Visit(movieListURL)
	})

	movieListScraper.OnHTML("a[href*=film]", func(e *colly.HTMLElement) {
		log.Println("Found movie link for ", strings.TrimSpace(e.Text))
		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		//movieScraper.Visit(movieURL)
	})

	movieScraper.OnHTML("a[href*=large]", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found image", movieImageURL)
	})

	(*scraper).Visit(url)

	// if visited, _ := (*scraper).HasVisited(url); !visited {
	// 	log.Println("not visited, lets go", url)
	// 	(*scraper).Visit(url)
	// } else {
	// 	log.Println("already visited", url)
	// }

}

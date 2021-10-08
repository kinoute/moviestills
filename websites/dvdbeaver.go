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

// This webpage stores a list of links to movie list pages sorted by alphabet (#, a, z).
// It's a good starting point for our task.
const BeaverURL string = "http://www.dvdbeaver.com/film/reviews.htm"

func DVDBeaverScraper(scraper **colly.Collector, options *config.Options) {

	log.Println("Starting DVDBeaver Scraper...")

	// Change allowed domain for the main scraper.
	// Since everything is served on the same domain, only one domain is necessary.
	(*scraper).AllowedDomains = []string{
		"www.dvdbeaver.com",
		"DVDBeaver.com",
		"www.DVDBeaver.com",
	}

	// The reviews page might have been updated so
	// we have to revisit it when restarting the scraper.
	// It is a single page, it will not cost anything anyway.
	(*scraper).AllowURLRevisit = true

	// Movies list might be updated often with new movies
	// so we autorize the scraper to revisit these pages.
	movieListScraper := (*scraper).Clone()
	movieListScraper.AllowURLRevisit = true

	// DVDBeaver pages have a weird charset.
	// Colly can deal with this automatically
	// and handle weird accents/characters better.
	movieListScraper.DetectCharset = true

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
		log.Println("visiting reviews page", r.URL.String())
	})

	// Find links to movies list by alphabet
	(*scraper).OnHTML("a[href*='listing' i]", func(e *colly.HTMLElement) {
		movieListURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found movie list page link", movieListURL)

		if err := movieListScraper.Visit(movieListURL); err != nil {
			log.Println("Can't visit movie list page:", err)
		}

		movieListScraper.Wait()
	})

	// Before making a request print "Visiting ..."
	movieListScraper.OnRequest(func(r *colly.Request) {
		log.Println("visiting movie list page", r.URL.String())
	})

	// Look for movie reviews links and create folder for every
	// movie we find to prepare the download of the snapshots.
	//
	// We have to iterate through each p element to discard
	// Blu-Ray reviews as we want to focus only on DVD reviews.
	movieListScraper.OnHTML("td p", func(e *colly.HTMLElement) {

		// We ignore BD reviews pages because we have
		// a specific scraper, "BluBeaver", for these
		// pages with Blu-Ray screenshots.
		if strings.Contains(e.DOM.Text(), "BD") || strings.Contains(e.DOM.Text(), "UHD") {
			log.Println("BD review, skipping")
			return
		}

		// Get the DVD movie review link. Sometimes there is
		// no link that matches our query so we stop right here.
		//
		// We use the CSS4 case-insensitive feature "i" to make sure
		// our filter will find everything, no matter the case.
		reviewLink := e.DOM.Find("a[href*='film' i]")
		movieURL, exists := reviewLink.Attr("href")
		if !exists {
			log.Println("no movie review link could be found, next")
			return
		}

		// Take care of weird characters in the movie's title
		movieName, err := utils.Normalize(reviewLink.Text())
		if err != nil || movieName == "" {
			log.Println("Can't normalize Movie name for", reviewLink.Text())
			return
		}

		log.Println("Found movie link for", movieName)

		// Create folder to save images in case it doesn't exist yet
		moviePath, err := utils.CreateFolder(options.DataDir, options.Website, movieName)
		if err != nil {
			log.Printf("Error creating folder for movie %v on %v: %v", movieName, options.Website, err)
			return
		}

		// Pass the movie path to the next request context
		// in order to save the images in the correct folder.
		ctx := colly.NewContext()
		ctx.Put("movie_path", moviePath)

		// Make sure we handle relative URLs if any
		movieURL = e.Request.AbsoluteURL(movieURL)

		log.Println("visiting movie page", movieURL)

		if err = movieScraper.Request("GET", movieURL, nil, ctx, nil); err != nil {
			log.Println("Can't visit movie page:", err)
		}
	})

	// Look for links on images that redirects to a "largest" version.
	// It is unlikely to find some of these on some DVD reviews, but sometimes
	// they compare DVD releases with BD releases and provide some images
	// with native resolution (1080p).
	//
	// We try to avoid images with "subs" in the filename as they are
	// most likely images with subtitles on top. We don't want that.
	movieScraper.OnHTML("a[href*='large' i]:not([href*='subs' i])", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Found linked image", movieImageURL)

		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Println("Can't request linked image:", err)
		}
	})

	// On DVD reviews, there are almost never clickable large versions.
	// Therefore we download the images as shown on the webpage and
	// be sure we avoid some weird ones (subtitles, DVD covers etc).
	movieScraper.OnHTML(
		"img:not([src*='banner' i])"+
			":not([src*='bitrate' i])"+
			":not([src$='gif' i])"+
			":not([src*='subs' i])"+
			":not([src*='daggers' i])"+
			":not([src*='posters' i])"+
			":not([src*='title' i])"+
			":not([src*='menu' i])", func(e *colly.HTMLElement) {
			movieImageURL := e.Request.AbsoluteURL(e.Attr("src"))

			// Filter low resolutions images to avoid false positives.
			// if the images are too small, we won't be able to use them
			// anyway so let's skip them.
			movieImageWidth, _ := strconv.Atoi(e.Attr("width"))
			movieImageHeight, _ := strconv.Atoi(e.Attr("height"))

			if movieImageHeight >= 275 && movieImageWidth >= 500 {
				log.Println("Image resolution seems OK, downloading", movieImageURL)
				if err := e.Request.Visit(movieImageURL); err != nil {
					log.Println("Can't request linked image:", err)
				}
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
					log.Println("Can't save image:", err)
				}
			}

			return
		}
	})

	if err := (*scraper).Visit(BeaverURL); err != nil {
		log.Println("Can't visit index page:", err)
	}

	// In case we enabled asynchronous jobs
	(*scraper).Wait()
}

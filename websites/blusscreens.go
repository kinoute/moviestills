package websites

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"moviestills/utils"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

// We use the cinematographers page because the movie's title
// can be easily scrapped/parsed there instead of the screen
// captures page (#, A, Z) where links are images.
//
// The downside is, some movies (animation?) might be missing
// because they don't have cinematographers associated to it.
const BlusURL string = "https://www.bluscreens.net/cinematographers.html"

// don't download images that are less than 20kb.
// it is helpful for this website since all images are hosted on
// imgur and some might have been deleted. When an image is deleted
// on imgur, it returns a small image with some text on it. We don't want that
const MinimumSize int = 1024 * 20

// data structure to save our result
// in our JSON file
type BlusMovie struct {
	Title string
	IMDb  string
	Path  string
}

func BlusScraper(scraper **colly.Collector) {

	log.Println("Starting Blus Screens Scraper...")

	// save movie infos as JSON
	movies := make([]*BlusMovie, 0)

	// change allowed domains for the main scrapper
	// images are stored either on imgur or postimage
	// so make sure we allow all these domains
	(*scraper).AllowedDomains = []string{
		"www.bluscreens.net",
		"imgur.com",
		"i.imgur.com",
		"postimage.org",
		"postimg.cc",
		"i.postimg.cc",
	}

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
	(*scraper).OnHTML("h2.wsite-content-title a[href*=html][href*=oss]", func(e *colly.HTMLElement) {

		movieName, err := utils.Normalize(e.Text)
		if err != nil || movieName == "" {
			log.Println("Can't normalize Movie name for", movieName)
			return
		}

		log.Println("Found movie link for", movieName)

		// create folder to save images in case it doesn't exist
		moviePath := filepath.Join(".", "data", "blusscreens", movieName)
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

		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("visiting movie page", movieURL)

		movieScraper.Request("GET", movieURL, nil, ctx, nil)
	})

	// go through each link to imgur found on the movie page
	movieScraper.OnHTML("div.galleryInnerImageHolder a[href*=imgur]", func(e *colly.HTMLElement) {

		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("inside movie page for", e.Request.Ctx.Get("movie_name"))

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

	// some old pages of blusscreens have a different layout.
	// we need a special funtion to handle this
	// eg: https://www.bluscreens.net/oss-117-rio-ne-reacutepond-plus.html
	movieScraper.OnHTML("div.galleryInnerImageHolder a[href*=postimage]", func(e *colly.HTMLElement) {
		postImgURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("found postimage link", postImgURL)

		e.Request.Visit(postImgURL)
	})

	// get full images from postimage.cc host
	// we need do get the "download" button link as
	// the image shown on the page is in "low" resolution
	movieScraper.OnHTML("div#content a#download[href*=postimg]", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("found postimg image", movieImageURL)

		e.Request.Visit(movieImageURL)
	})

	// get IMDB id from the IMDB link
	// it is an unique ID for each movie and it is
	// better to use it than the movie's title.
	movieScraper.OnHTML("div.wsite-image-border-none a[href*=imdb]", func(e *colly.HTMLElement) {
		imdbLink := e.Attr("href")

		log.Println("found imdb link", e.Attr("href"))

		// isolate IMDB id from IMDb url
		re := regexp.MustCompile(`(tt\d{7,8})`)
		imdbID := re.FindString(imdbLink)

		// now we have everything to append this movie to our JSON results
		movie := &BlusMovie{
			Title: e.Request.Ctx.Get("movie_name"),
			IMDb:  imdbID,
			Path:  e.Request.Ctx.Get("movie_path"),
		}

		movies = append(movies, movie)

		// we save the JSON results after every movie
		// in case we have to stop the scrapping in
		// the middle. At least, we will have the
		// intermediate datas.
		writeJSON(movies)
	})

	// check what we just visited and if its an image
	// save it to the movie folder we created earlier
	movieScraper.OnResponse(func(r *colly.Response) {

		// if we're dealing with an image, save it in the correct folder
		if strings.Index(r.Headers.Get("Content-Type"), "image") > -1 {

			// ignore weird small-sized images
			imageSize, _ := strconv.Atoi(r.Headers.Get("Content-Length"))

			if imageSize < MinimumSize {
				log.Println("Small-sized image, not downloading", r.FileName())
				return
			}

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

// func to save summary of scrapped movies to json
func writeJSON(data []*BlusMovie) {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Println("Unable to create Blus Screens json file")
		return
	}

	_ = ioutil.WriteFile("blusscreens.json", file, 0644)
}

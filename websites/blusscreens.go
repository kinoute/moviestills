package websites

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

// We use the cinematographers page because the movie's title
// can be easily scrapped/parsed there instead of the screen
// captures page (#, A, Z) where links are images.
//
// The downside is, some movies (animation?) might be missing
// because they don't have cinematographers associated to it.
var BlusURL = "https://www.bluscreens.net/cinematographers.html"

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

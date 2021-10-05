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

// Don't download images that are less than 20kb.
// It is helpful for this website since most of the images are hosted
// on imgur and some might have been deleted. When an image is deleted
// on imgur, it returns a small image with some text on it. We don't want that.
const MinimumSize int = 1024 * 20

// Data structure to save our result in our JSON file
type BlusMovie struct {
	Title string
	IMDb  string
	Path  string
}

func BlusScraper(scraper **colly.Collector) {

	log.Println("Starting Blus Screens Scraper...")

	// Save movie infos as JSON
	movies := make([]*BlusMovie, 0)

	// Change allowed domains for the main scrapper.
	// Images are stored either on imgur or postimage
	// so make sure we allow all these domains.
	(*scraper).AllowedDomains = []string{
		"www.bluscreens.net",
		"imgur.com",
		"i.imgur.com",
		"postimage.org",
		"postimg.cc",
		"i.postimg.cc",
		"pixxxels.cc",
		"i.pixxxels.cc",
	}

	// The cinematographers page might have been updated so
	// we have to revisit it when restarting scraper.
	(*scraper).AllowURLRevisit = true

	// Scraper to fetch movie images.
	// Movie pages are not updated after being published therefore we only visit once.
	movieScraper := (*scraper).Clone()
	movieScraper.AllowURLRevisit = false

	// Before making a request print "Visiting ..."
	(*scraper).OnRequest(func(r *colly.Request) {
		log.Println("visiting movie list page", r.URL.String())
	})

	// Isolate every movie listed, keep its title and
	// create a dedicated folder if it doesn't exist
	// to store images.
	//
	// Then visit movie page where images are listed/displayed.
	(*scraper).OnHTML("h2.wsite-content-title a[href*=html]", func(e *colly.HTMLElement) {

		movieName, err := utils.Normalize(e.Text)
		if err != nil || movieName == "" {
			log.Println("Can't normalize Movie name for", e.Text)
			return
		}

		log.Println("Found movie link for", movieName)

		// Create folder to save images in case it doesn't exist
		moviePath := filepath.Join(".", "data", "blusscreens", movieName)
		if err = os.MkdirAll(moviePath, os.ModePerm); err != nil {
			log.Println("Error creating folder for", movieName)
			return
		}

		// Pass the movie's name and path to the next request context
		// in order to save the images in correct folder.
		ctx := colly.NewContext()
		ctx.Put("movie_name", movieName)
		ctx.Put("movie_path", moviePath)

		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("visiting movie page", movieURL)

		if err = movieScraper.Request("GET", movieURL, nil, ctx, nil); err != nil {
			log.Println("Can't request movie page:", err)
		}

		// In case we enabled asynchronous jobs
		movieScraper.Wait()
	})

	// Go through each link to imgur found on the movie page
	movieScraper.OnHTML(
		"div.galleryInnerImageHolder a[href*=imgur], "+
			"td.wsite-multicol-col div a[href*=imgur]", func(e *colly.HTMLElement) {

			movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
			log.Println("inside movie page for", e.Request.Ctx.Get("movie_name"))

			// Create link to the real image if it's a link to imgur's
			// website and not directly to the image.
			// eg. https://imgur.com/ABC to https://i.imgur.com/ABC.png
			if !strings.Contains(movieImageURL, "i.imgur.com") {
				movieImageURL += ".png"
				movieImageURL = strings.Replace(movieImageURL, "https://imgur.com", "https://i.imgur.com", 1)
			}

			log.Println("Found linked image", movieImageURL)
			if err := e.Request.Visit(movieImageURL); err != nil {
				log.Println("Can't request linked image:", err)
			}
		})

	// Some old pages of blusscreens have a different layout.
	// We need a special funtion to handle this.
	// eg: https://www.bluscreens.net/oss-117-rio-ne-reacutepond-plus.html
	movieScraper.OnHTML("div.galleryInnerImageHolder a[href*=postimage]", func(e *colly.HTMLElement) {
		postImgURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("found postimage link", postImgURL)

		if err := e.Request.Visit(postImgURL); err != nil {
			log.Println("Can't request linked image:", err)
		}
	})

	// Another kind of weird layout mixing table and div.
	// eg: https://www.bluscreens.net/pain--gain.html
	movieScraper.OnHTML("td.wsite-multicol-col div a[href*=postim]", func(e *colly.HTMLElement) {
		postImgURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("found postimage.org link", postImgURL)

		// Some links redirect to "postimg.org" and later "pixxxels.cc".
		// "postimg.org" is not available anymore, we might need to rewrite the URLs.
		postImgURL = strings.Replace(postImgURL, "postimg.org", "postimage.org", 1)

		if err := e.Request.Visit(postImgURL); err != nil {
			log.Println("Can't request linked image:", err)
		}
	})

	// Get full images from postimage.cc host.
	// We need to get the "download" button link as
	// the image shown on the page is in a "lower" resolution.
	movieScraper.OnHTML(
		"div#content a#download[href*=postimg], "+
			"div#content a#download[href*=pixxxels]", func(e *colly.HTMLElement) {
			movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
			log.Println("found postimg image", movieImageURL)

			if err := e.Request.Visit(movieImageURL); err != nil {
				log.Println("Can't request linked image:", err)
			}
		})

	// Get IMDB id from the IMDB link.
	// It is an unique ID for each movie and it is
	// better to use it than the movie's title.
	movieScraper.OnHTML("div.wsite-image-border-none a[href*=imdb]", func(e *colly.HTMLElement) {
		imdbLink := e.Attr("href")

		log.Println("found imdb link", imdbLink)

		// Isolate IMDB id from IMDb url
		re := regexp.MustCompile(`(tt\d{7,8})`)
		imdbID := re.FindString(imdbLink)

		// Now we have everything to append this movie to our JSON results
		movie := &BlusMovie{
			Title: e.Request.Ctx.Get("movie_name"),
			IMDb:  imdbID,
			Path:  e.Request.Ctx.Get("movie_path"),
		}

		movies = append(movies, movie)

		// We save the JSON results after every movie
		// in case we have to stop the scrapping in
		// the middle. At least, we will have some
		// intermediate datas.
		writeJSON(movies)
	})

	// Check what we just visited and if its an image
	// save it to the movie folder we created earlier.
	movieScraper.OnResponse(func(r *colly.Response) {

		// If we're dealing with an image, save it in the correct folder
		if strings.Contains(r.Headers.Get("Content-Type"), "image") {

			// Ignore weird small-sized images
			imageSize, err := strconv.Atoi(r.Headers.Get("Content-Length"))
			if err != nil {
				log.Println("Can't get image size from headers:", err)
				return
			}

			if imageSize < MinimumSize {
				log.Println("Small-sized image, not downloading", r.FileName())
				return
			}

			outputDir := r.Ctx.Get("movie_path")
			outputImgPath := outputDir + "/" + r.FileName()

			// Save only if we don't already downloaded it
			if _, err := os.Stat(outputImgPath); os.IsNotExist(err) {
				if err = r.Save(outputImgPath); err != nil {
					log.Println("Can't save image:", err)
				}
			}

			return
		}
	})

	if err := (*scraper).Visit(BlusURL); err != nil {
		log.Println("Can't visit index page:", err)
	}

	// In case we enabled asynchronous jobs
	(*scraper).Wait()

}

// Save summary of scrapped movies to a JSON file
func writeJSON(data []*BlusMovie) {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Println("Unable to create Blus Screens json file:", err)
		return
	}

	_ = ioutil.WriteFile("blusscreens.json", file, 0644)
}

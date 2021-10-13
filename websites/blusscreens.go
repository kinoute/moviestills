package websites

import (
	"moviestills/config"
	"moviestills/utils"
	"strconv"
	"strings"

	log "github.com/pterm/pterm"

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
// type BlusMovie struct {
// 	Title string
// 	IMDb  string
// 	Path  string
// }

func BlusScraper(scraper **colly.Collector, options *config.Options) {

	// Save movie infos as JSON
	// movies := make([]*BlusMovie, 0)

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
	// we have to revisit it when restarting the scraper.
	(*scraper).AllowURLRevisit = true

	// Scraper to fetch movie images.
	// Movie pages are not updated after being published
	// therefore we only visit them once.
	movieScraper := (*scraper).Clone()
	movieScraper.AllowURLRevisit = false

	// Print error just in case
	(*scraper).OnError(func(r *colly.Response, err error) {
		log.Error.Println(r.Request.URL, "\t", log.White(r.StatusCode), "\nError:", log.Red(err))
	})

	// Before making a request print "Visiting ..."
	(*scraper).OnRequest(func(r *colly.Request) {
		log.Debug.Println("visiting index page", log.White(r.URL.String()))
	})

	// Isolate every movie listed, keep its title and
	// create a dedicated folder if it doesn't exist
	// to store images.
	//
	// Then visit movie page where images are listed/displayed.
	(*scraper).OnHTML("h2.wsite-content-title a[href*=html]", func(e *colly.HTMLElement) {

		// Remove weird accents and spaces from the movie's title
		movieName, err := utils.Normalize(e.Text)
		if err != nil || movieName == "" {
			log.Error.Println("Can't normalize Movie name for", log.White(e.Text), ":", log.Red(err))
			return
		}

		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug.Println("Found movie page link", log.White(movieURL))

		// Create folder to save images in case it doesn't exist
		moviePath, err := utils.CreateFolder(options.DataDir, options.Website, movieName)
		if err != nil {
			log.Error.Println("Can't create movie folder for:", log.White(movieName), log.Red(err))
			return
		}

		log.Info.Println("Found movie page for:", log.White(movieName))

		// Pass the movie's name and path to the next request context
		// in order to save the images in correct folder.
		ctx := colly.NewContext()
		ctx.Put("movie_name", movieName)
		ctx.Put("movie_path", moviePath)

		// Try to visit movie page
		if err = movieScraper.Request("GET", movieURL, nil, ctx, nil); err != nil {
			log.Error.Println("Can't get movie page", log.White(movieURL), ":", log.Red(err))
		}

		// In case we enabled asynchronous jobs
		movieScraper.Wait()
	})

	// Go through each link to imgur found on the movie page
	movieScraper.OnHTML(
		"div.galleryInnerImageHolder a[href*=imgur], "+
			"td.wsite-multicol-col div a[href*=imgur]", func(e *colly.HTMLElement) {

			movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
			log.Debug.Println("inside movie page for", log.White(e.Request.Ctx.Get("movie_name")))

			// Create link to the real image if it's a link to imgur's
			// website and not directly to the image.
			// eg. https://imgur.com/ABC to https://i.imgur.com/ABC.png
			if !strings.Contains(movieImageURL, "i.imgur.com") {
				movieImageURL += ".png"
				movieImageURL = strings.Replace(movieImageURL, "https://imgur.com", "https://i.imgur.com", 1)
			}

			log.Debug.Println("Found linked image", log.White(movieImageURL))
			if err := e.Request.Visit(movieImageURL); err != nil {
				log.Error.Println("Can't get linked image", log.White(movieImageURL), ":", log.Red(err))
			}
		})

	// Some old pages of blusscreens have a different layout.
	// We need a special funtion to handle this.
	// eg: https://www.bluscreens.net/oss-117-rio-ne-reacutepond-plus.html
	movieScraper.OnHTML("div.galleryInnerImageHolder a[href*=postimage]", func(e *colly.HTMLElement) {
		postImgURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug.Println("found postimage link", log.White(postImgURL))

		if err := e.Request.Visit(postImgURL); err != nil {
			log.Error.Println("Can't request postimage link", log.White(postImgURL), ":", log.Red(err))
		}
	})

	// Another kind of weird layout mixing table and div.
	// eg: https://www.bluscreens.net/pain--gain.html
	movieScraper.OnHTML("td.wsite-multicol-col div a[href*=postim]", func(e *colly.HTMLElement) {
		postImgURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug.Println("found postimage.org link", log.White(postImgURL))

		// Some links redirect to "postimg.org" and later "pixxxels.cc".
		// "postimg.org" is not available anymore, we might need to rewrite the URLs.
		postImgURL = strings.Replace(postImgURL, "postimg.org", "postimage.org", 1)

		if err := e.Request.Visit(postImgURL); err != nil {
			log.Error.Println("Can't request postimage link", log.White(postImgURL), ":", log.Red(err))
		}
	})

	// Get full images from postimage.cc host.
	// We need to get the "download" button link as
	// the image shown on the page is in a "lower" resolution.
	movieScraper.OnHTML(
		"div#content a#download[href*=postimg], "+
			"div#content a#download[href*=pixxxels]", func(e *colly.HTMLElement) {
			movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
			log.Debug.Println("found postimg full image", log.White(movieImageURL))

			if err := e.Request.Visit(movieImageURL); err != nil {
				log.Error.Println("Can't get postimage full image", log.White(movieImageURL), ":", log.Red(err))
			}
		})

	// Get IMDB id from the IMDB link.
	// It is an unique ID for each movie and it is
	// better to use it than the movie's title.
	// movieScraper.OnHTML("div.wsite-image-border-none a[href*=imdb]", func(e *colly.HTMLElement) {
	// 	imdbLink := e.Attr("href")

	// 	log.Info.Println("found IMDb link", log.White(imdbLink))

	// 	// Isolate IMDB id from IMDb url
	// 	re := regexp.MustCompile(`(tt\d{7,8})`)
	// 	imdbID := re.FindString(imdbLink)

	// 	// Now we have everything to append this movie to our JSON results
	// 	movie := &BlusMovie{
	// 		Title: e.Request.Ctx.Get("movie_name"),
	// 		IMDb:  imdbID,
	// 		Path:  e.Request.Ctx.Get("movie_path"),
	// 	}

	// 	movies = append(movies, movie)

	// 	// We save the JSON results after every movie
	// 	// in case we have to stop the scrapping in
	// 	// the middle. At least, we will have some
	// 	// intermediate datas.
	// 	writeJSON(movies)
	// })

	// Check what we just visited and if its an image
	// save it to the movie folder we created earlier.
	movieScraper.OnResponse(func(r *colly.Response) {

		// Ignore anything that is not an image
		if !strings.Contains(r.Headers.Get("Content-Type"), "image") {
			return
		}

		// Calculate Image Size from Headers
		imageSize, err := strconv.Atoi(r.Headers.Get("Content-Length"))
		if err != nil {
			log.Error.Println("Can't get image size from headers:", log.Red(err))
			return
		}

		// Images are hosted on imgur and some might have been deleted. When an image is deleted
		// on imgur, it returns a small image with some text on it. We don't want that.
		if imageSize < MinimumSize {
			log.Error.Println("Small-sized image, not downloading", log.White(r.FileName()))
			return
		}

		// Try to save movie image
		if err := utils.SaveImage(r.Ctx.Get("movie_path"),
			r.Ctx.Get("movie_name"),
			r.FileName(),
			r.Body,
			options.Hash,
		); err != nil {
			log.Error.Println("Can't save image", log.White(r.FileName()), log.Red(err))
		}
	})

	if err := (*scraper).Visit(BlusURL); err != nil {
		log.Error.Println("Can't visit index page", log.White(BlusURL), log.Red(err))
	}

	// In case we enabled asynchronous jobs
	(*scraper).Wait()

}

// Save summary of scrapped movies to a JSON file
// func writeJSON(data []*BlusMovie) {
// 	file, err := json.MarshalIndent(data, "", " ")
// 	if err != nil {
// 		log.Println("Unable to create Blus Screens json file:", err)
// 		return
// 	}

// 	_ = ioutil.WriteFile("blusscreens.json", file, 0644)
// }

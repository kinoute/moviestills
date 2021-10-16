package websites

import (
	"moviestills/config"
	"moviestills/utils"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	log "github.com/pterm/pterm"
)

// This webpage stores a list of links to movie reviews of Blu-rays
const BluBeaverURL string = "http://www.dvdbeaver.com/blu-ray.htm"

// Main function that handles all the scraping logic for this website
func BluBeaverScraper(scraper **colly.Collector, options *config.Options) {

	// Change allowed domain for the main scraper.
	// Since everything is served on the same domain,
	// only one domain is really necessary.
	(*scraper).AllowedDomains = []string{
		"www.blubeaver.ca",
		"www.dvdbeaver.com",
		"dvdbeaver.com",
		"DVDBeaver.com",
		"www.DVDBeaver.com",
	}

	// The index page might have been updated since last visit so
	// we have to revisit it when restarting the scraper.
	// It is a single page, it will not cost anything anyway.
	(*scraper).AllowURLRevisit = true

	// BluBeaver pages have a weird charset.
	// Colly can deal with this automatically
	// and handle weird accents/characters better.
	(*scraper).DetectCharset = true

	// Scraper to fetch movie images on reviews pages.
	// These pages are not updated after being
	// published therefore we only visit them once.
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

	// Find links to movies reviews and isolate the movie's title.
	// Since BluBeaver is somewhat a custom website, some links
	// might have different cases. We use the CSS4 "i" case-insensitive
	// feature to make sure our filter doesn't miss anything.
	(*scraper).OnHTML("li a[href*='film' i][href$='htm' i]", func(e *colly.HTMLElement) {

		// Sometimes, Blubeaver made mistakes and added links to reviews
		// on Amazon icons. Since we use the link to isolate the movie's title,
		// we ignore these links as they don't have the movie's name included.
		if _, iconExistsInLink := e.DOM.Find("img").Attr("src"); iconExistsInLink {
			log.Debug.Println("Link without text, just an icon, next")
			return
		}

		movieURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug.Println("Found movie page link", log.White(movieURL))

		// Remove weird accents and spaces from the movie's title
		movieName, err := utils.Normalize(e.Text)
		if err != nil {
			log.Error.Println("Can't normalize Movie name for", log.White(e.Text), ":", log.Red(err))
			return
		}

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

	// Before making a request to URL
	movieScraper.OnRequest(func(r *colly.Request) {
		log.Debug.Println("visiting", log.White(r.URL.String()))
	})

	// It's rare but sometimes on BD reviews there are no large versions.
	// Therefore we download the images as shown on the webpage and
	// be sure we avoid some weird ones (subtitles, DVD covers etc).
	movieScraper.OnHTML(
		":not(a) >"+
			"img:not([src*='banner' i])"+
			":not([src*='rating' i])"+
			":not([src*='package' i])"+
			":not([src*='bitrate' i])"+
			":not([src*='bitgraph' i])"+
			":not([src$='gif' i])"+
			":not([src*='sub' i])"+
			":not([src*='daggers' i])"+
			":not([src*='poster' i])"+
			":not([src*='title' i])"+
			":not([src*='menu' i])", func(e *colly.HTMLElement) {
			movieImageURL := e.Request.AbsoluteURL(e.Attr("src"))

			// Filter low resolutions images to avoid false positives.
			// if the images are too small, we won't be able to use them
			// anyway so let's skip them.
			movieImageWidth, _ := strconv.Atoi(e.Attr("width"))
			movieImageHeight, _ := strconv.Atoi(e.Attr("height"))

			if movieImageHeight >= 265 && movieImageWidth >= 500 {
				if err := e.Request.Visit(movieImageURL); err != nil {
					log.Error.Println("Can't get inline image", log.White(movieImageURL), ":", log.Red(err))
				}
			}
		})

	// Look for links on images that redirects to a "largest" version.
	// These links appear on Blu-Ray reviews almost exclusively and
	// provide images with native resolution (1080p).
	//
	// We try to avoid images with "subs" in the filename as they are
	// most likely images with subtitles on top. We don't want that.
	movieScraper.OnHTML("a[href*='large' i]:not([href*='subs' i])", func(e *colly.HTMLElement) {
		movieImageURL := e.Request.AbsoluteURL(e.Attr("href"))
		log.Debug.Println("Found large image", log.White(movieImageURL))

		if err := e.Request.Visit(movieImageURL); err != nil {
			log.Error.Println("Can't get large image", log.White(movieImageURL), ":", log.Red(err))

			// Sometimes, the high quality version of an image
			// is not available anymore ("Not Found").
			//
			// In this case, we can try to save the image
			// shown on the webpage that has a lower resolution.
			lowImageURL, imgExists := e.DOM.Find("img").Attr("src")
			if !imgExists {
				log.Error.Println("Could not find an image inside link", log.White(movieImageURL))
				return
			}

			log.Info.Println("Trying to save low quality image instead", log.White(lowImageURL))
			if err := e.Request.Visit(e.Request.AbsoluteURL(lowImageURL)); err != nil {
				log.Error.Println("Can't get low resolution image", log.White(lowImageURL), ":", log.Red(err))
			}
		}
	})

	// Check if what we just visited is an image and
	// save it to the movie folder we created earlier.
	movieScraper.OnResponse(func(r *colly.Response) {

		// Ignore anything that is not an image
		if !strings.Contains(r.Headers.Get("Content-Type"), "image") {
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

	if err := (*scraper).Visit(BluBeaverURL); err != nil {
		log.Error.Println("Can't visit index page", log.White(BluBeaverURL), ":", log.Red(err))
	}

	// In case we enabled asynchronous jobs
	(*scraper).Wait()
}

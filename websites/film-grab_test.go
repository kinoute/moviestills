package websites

import (
	"log"
	"moviestills/utils"
	"testing"
)

// Test Film-grab index page
func TestFilmGrabIndexPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode(FilmGrabURL)

	// Number of entries
	numMovies := doc.Find("div#primary a.title[href*=film]").Length()
	if numMovies < 2500 {
		log.Fatalln("Number of movie reviews seem really low", numMovies)
	}
}

// 12 Angry Men
func TestFilmGrabNormalMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://film-grab.com/2013/11/19/12-angry-men/")

	// We should find many links to high-quality images
	numLargeImages := doc.Find("div.bwg_container div.bwg-item a.bwg-a[href*=film]").Length()
	if numLargeImages != 65 {
		log.Fatalln("Number of links to large images is different than 65:", numLargeImages)
	}

}

package websites

import (
	"log"
	"moviestills/utils"
	"strconv"
	"testing"
)

// Test Film-grab index page
func TestMovieScreencapsIndexPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode(ScreenCapsURL)

	// Number of entries
	numMovies := doc.Find("div.tagindex ul.links li a[href*=movie]").Length()
	if numMovies < 700 {
		log.Fatalln("Number of movie reviews seem really low", numMovies)
	}
}

// Eagle Eye
func TestMovieScreencapsNormalMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://movie-screencaps.com/eagle-eye-2008/")

	// We should find many links to high-quality images
	numLargeImages := doc.Find("section.entry-content a[href*=wp][href*=caps]").Length()
	if numLargeImages != 180 {
		log.Fatalln("Number of links to large images is different than 180:", numLargeImages)
	}

	// Get number of pages
	numPages, exists := doc.Find("div.pixcode + div.wp-pagenavi > select.paginate option:last-child").Attr("value")
	if !exists {
		log.Fatalln("Can't retrieve number of pages")
	}

	numPagesInt, err := strconv.Atoi(numPages)
	if err != nil {
		log.Fatalln("Can't convert number of pages to int", err)
	}

	if numPagesInt != 74 {
		log.Fatalln("Number of pages is different than 74:", numPagesInt)
	}

}

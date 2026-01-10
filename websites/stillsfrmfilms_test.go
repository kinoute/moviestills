package websites

import (
	"moviestills/utils"
	"testing"
)

// Test StillsFrmFilms index page
func TestStillsFrmIndexPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode(StillsFrmFilmsURL)

	// Number of entries
	numMovies := doc.Find("div.page-body div.wp-caption").Length()
	if numMovies < 50 {
		t.Fatalf("Number of movie reviews seem really low: %d", numMovies)
	}
}

// 25th Hour
func TestStillsFrmNormalMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://stillsfrmfilms.wordpress.com/2012/09/17/25th-hour/")

	// We should find many links to high-quality images
	numLargeImages := doc.Find("div.photo-inner dl.gallery-item a[href*=stills] img[src*=uploads]").Length()
	if numLargeImages != 55 {
		t.Fatalf("Number of links to large images is different than 55: %d", numLargeImages)
	}
}

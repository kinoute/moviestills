package websites

import (
	"moviestills/utils"
	"strconv"
	"testing"
)

// Test movie-screencaps index page
func TestMovieScreencapsIndexPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode(ScreenCapsURL)

	// Number of entries
	numMovies := doc.Find("div.tagindex ul.links li a[href*=movie]").Length()
	if numMovies < 700 {
		t.Fatalf("Number of movie reviews seem really low: %d", numMovies)
	}
}

// Eagle Eye
func TestMovieScreencapsNormalMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://movie-screencaps.com/eagle-eye-2008/")

	// We should find many links to high-quality images
	numLargeImages := doc.Find("section.entry-content a[href*=jpg]").Length()
	if numLargeImages != 180 {
		t.Fatalf("Number of links to large images is different than 180: %d", numLargeImages)
	}

	// Get number of pages
	numPages, exists := doc.Find("div.pixcode + div.wp-pagenavi > select.paginate option:last-child").Attr("value")
	if !exists {
		t.Fatal("Can't retrieve number of pages")
	}

	numPagesInt, err := strconv.Atoi(numPages)
	if err != nil {
		t.Fatalf("Can't convert number of pages to int: %v", err)
	}

	if numPagesInt != 74 {
		t.Fatalf("Number of pages is different than 74: %d", numPagesInt)
	}
}

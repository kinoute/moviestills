package websites

import (
	"moviestills/utils"
	"testing"
)

// Test EvanERichards cinematographer page
func TestEvanIndexPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode(EvanERichardsURL)

	// Number of entries (including TV series etc)
	numMovies := doc.Find("tbody tr.pp-table-row").Length()
	if numMovies < 250 {
		t.Fatalf("Number of movie reviews seem really low: %d", numMovies)
	}
}

// 12 Monkeys
func TestEvanNormalMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://www.evanerichards.com/2009/28")

	// We should find many links to high-quality images
	numLargeImages := doc.Find("div.elementor-widget-container div.ngg-gallery-thumbnail a[class*=shutter]").Length()
	if numLargeImages != 351 {
		t.Fatalf("Number of links to large images is different than 351: %d", numLargeImages)
	}
}

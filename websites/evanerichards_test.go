package websites

import (
	"moviestills/utils"
	"testing"
)

// Test EvanERichards index page
func TestEvanIndexPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode(EvanERichardsURL)

	// The site now uses plain table rows without the pp-table-row class
	// Count all table rows that likely contain movie entries
	numRows := doc.Find("tbody tr").Length()
	if numRows < 250 {
		t.Fatalf("Number of table rows seem really low: %d", numRows)
	}
}

// 12 Monkeys
func TestEvanNormalMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://www.evanerichards.com/2009/28")

	// We should find many links to high-quality images
	numLargeImages := doc.Find("div.elementor-widget-container div.ngg-gallery-thumbnail a[class*=shutter]").Length()
	if numLargeImages < 100 {
		t.Fatalf("Number of links to large images seems really low: %d", numLargeImages)
	}
}

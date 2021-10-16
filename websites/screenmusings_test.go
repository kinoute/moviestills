package websites

import (
	"log"
	"moviestills/utils"
	"testing"
)

// Test ScreenMusings index page
func TestScreenMusingsIndexPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode(ScreenMusingsURL)

	// Number of entries
	numMovies := doc.Find("nav#movies ul li a[href*=dvd], nav#movies ul li a[href*=blu]").Length()
	if numMovies < 200 {
		log.Fatalln("Number of movie reviews seem really low", numMovies)
	}
}

// Check if "most viewed stills" link can be found on movie page
// Annihilation
func TestScreenMusingsMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://screenmusings.org/movie/blu-ray/Annihilation/")

	mostLink, exists := doc.Find("ul#gallery-nav-top li:nth-last-child(2) a[href*=most]").Attr("href")
	if !exists {
		log.Fatalln("Most viewed page link can't be found")
	}

	if mostLink != "most-viewed-stills.htm" {
		log.Fatalln("The most viewed page is not good", mostLink)
	}

}

// Annihilation
func TestScreenMusingsMovieMostViewedPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://screenmusings.org/movie/blu-ray/Annihilation/most-viewed-stills.htm")

	// We should find many links to high-quality images
	numLargeImages := doc.Find("div#thumbnails div.thumb img[src*=thumb]").Length()
	if numLargeImages != 296 {
		log.Fatalln("Number of links to large images is different than 296:", numLargeImages)
	}

}

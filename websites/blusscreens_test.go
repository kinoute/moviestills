package websites

import (
	"log"
	"moviestills/utils"
	"testing"
)

// Test BlusScreens cinematographer page
func TestBlusIndexPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode(BlusURL)

	// Number of BD reviews listed on the index page
	numMovies := doc.Find("h2.wsite-content-title a[href*=html]").Length()
	if numMovies < 400 {
		log.Fatalln("Number of movie reviews seem really low", numMovies)
	}
}

// The "Giant" movie
// Normal movie review with large images available as links to imgur
func TestBlusNormalMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://www.bluscreens.net/giant.html")

	// We should find 50 links to imgur high-quality images
	numLargeImages := doc.Find("div.galleryInnerImageHolder a[href*=imgur], td.wsite-multicol-col div a[href*=imgur]").Length()
	if numLargeImages != 50 {
		log.Fatalln("Number of links to large images is different than 50:", numLargeImages)
	}

}

// Some old pages of blusscreens have a different layout.
// We need a special function to handle this.
// eg: OSS 117
func TestBlusAlternativeMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://www.bluscreens.net/skin-i-live-in-the.html")

	numLargeImages := doc.Find("div.galleryInnerImageHolder a[href*=postimage]").Length()
	if numLargeImages != 40 {
		log.Fatalln("Number of links to large images should be 40:", numLargeImages)
	}

}

// Another kind of weird layout mixing table and div.
// eg: Pain & Gain
func TestBlusAlternative2MoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://www.bluscreens.net/pain--gain.html")

	numLargeImages := doc.Find("td.wsite-multicol-col div a[href*=postim]").Length()
	if numLargeImages != 60 {
		log.Fatalln("Number of links to large images should be 60:", numLargeImages)
	}

}

// Get Link of full image from download button
func TestBlusDownloadImageLink(t *testing.T) {
	doc := utils.GetHTMLCode("https://pixxxels.cc/N2fk8KCs")

	imgURL, urlExists := doc.Find("div#content a#download[href*=postimg], div#content a#download[href*=pixxxels]").Attr("href")
	if !urlExists {
		log.Fatalln("Link of download image could not be found")
	}

	if imgURL != "https://i.postimg.cc/Q8y2cJC0/59.png?dl=1" {
		log.Fatalln("download image link not correct, expected", "https://i.postimg.cc/Q8y2cJC0/59.png?dl=1", "got", imgURL)
	}
}

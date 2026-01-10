package websites

import (
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
		t.Fatalf("Number of movie reviews seem really low: %d", numMovies)
	}
}

// The "Giant" movie
// Movie review with gallery link to imgur
func TestBlusNormalMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://www.bluscreens.net/giant.html")

	// The site now links to imgur/postimg galleries instead of individual images
	// Check that there's at least one gallery link
	numGalleryLinks := doc.Find("a[href*='imgur.com/a/'], a[href*='postimg.cc/gallery/']").Length()
	if numGalleryLinks < 1 {
		t.Fatalf("No gallery links found on page, expected at least 1: %d", numGalleryLinks)
	}
}

// Some pages of blusscreens link to postimg galleries.
// eg: The Skin I Live In
func TestBlusAlternativeMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://www.bluscreens.net/skin-i-live-in-the.html")

	// Check for gallery image holders (thumbnails displayed on page)
	numImageHolders := doc.Find("div.galleryInnerImageHolder").Length()
	if numImageHolders < 30 {
		t.Fatalf("Number of gallery image holders seems low: %d", numImageHolders)
	}
}

// Another layout using wsite-image elements in multicol tables.
// eg: Tie Me Up! Tie Me Down!
func TestBlusAlternative2MoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://www.bluscreens.net/tie-me-up-tie-me-down.html")

	// This page uses wsite-image divs instead of gallery holders
	numImages := doc.Find("div.wsite-image").Length()
	if numImages < 40 {
		t.Fatalf("Number of wsite-image elements seems low: %d", numImages)
	}
}

// Get Link to gallery from movie page
func TestBlusGalleryLink(t *testing.T) {
	doc := utils.GetHTMLCode("https://www.bluscreens.net/skin-i-live-in-the.html")

	_, urlExists := doc.Find("a[href*='postimg.cc/gallery/']").Attr("href")
	if !urlExists {
		t.Fatal("Gallery link to postimg.cc could not be found")
	}
}

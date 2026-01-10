package websites

import (
	"moviestills/utils"
	"strconv"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// Test number of movie lists index pages
func TestMainDVDBeaverPage(t *testing.T) {
	doc := utils.GetHTMLCode(BeaverURL)

	numPages := doc.Find("a[href*='listing' i]").Length()
	if numPages != 27 {
		t.Fatalf("Number of movie lists pages is bad: %d", numPages)
	}
}

// Test DVDBeaver "number/#" index page
func TestDVDBeaverIndexNumberPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("http://www.dvdbeaver.com/listing/num.htm")

	// Number of BD reviews listed on the index page
	numMovies := doc.Find("td p a[href*='film' i]").Length()
	if numMovies < 120 {
		t.Fatalf("Number of movie reviews seem really low: %d", numMovies)
	}
}

// The "3 Godfathers" movie
// Normal movie review with inline images
func TestDVDBeaverMoviePageWithOnlyInlineImages(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("http://www.dvdbeaver.com/film/DVDReviews19/3_godfathers_dvd_review.htm")

	// No large links found, should return zero
	numLargeImages := doc.Find("a[href*='large' i]:not([href*='subs' i])").Length()
	if numLargeImages != 0 {
		t.Fatalf("Number of large images should be zero: %d", numLargeImages)
	}

	// Only inline images available, should have 7 of them
	numInlineImgs := 0
	doc.Find(":not(a) >" +
		"img:not([src*='banner' i])" +
		":not([src*='rating' i])" +
		":not([src*='package' i])" +
		":not([src*='bitrate' i])" +
		":not([src*='bitgraph' i])" +
		":not([src$='gif' i])" +
		":not([src*='sub' i])" +
		":not([src*='daggers' i])" +
		":not([src*='poster' i])" +
		":not([src*='title' i])" +
		":not([src*='menu' i])").Each(func(i int, s *goquery.Selection) {
		height, _ := s.Attr("height")
		heightInt, _ := strconv.Atoi(height)
		width, _ := s.Attr("width")
		widthInt, _ := strconv.Atoi(width)
		if heightInt >= 265 && widthInt >= 500 {
			numInlineImgs++
		}
	})

	if numInlineImgs != 7 {
		t.Fatalf("Valid inline images should be 7, but found %d", numInlineImgs)
	}
}

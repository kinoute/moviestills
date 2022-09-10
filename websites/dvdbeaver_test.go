package websites

import (
	"log"
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
		log.Fatalln("Number of movie lists pages is bad", numPages)
	}
}

// Test DVDBeaver "number/#" index page
func TestDVDBeaverIndexNumberPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("http://www.dvdbeaver.com/listing/num.htm")

	// Number of BD reviews listed on the index page
	numMovies := doc.Find("td p a[href*='film' i]").Length()
	if numMovies < 120 {
		log.Fatalln("Number of movie reviews seem really low", numMovies)
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
		log.Fatalln("Number of large images should be zero:", numLargeImages)
	}

	// Only inline images available, should have 12 of them
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
			numInlineImgs += 1
		}

	})

	if numInlineImgs != 7 {
		log.Fatalln("Valid inline images should be 7, but found", numInlineImgs)
	}

}

package websites

import (
	"log"
	"moviestills/utils"
	"strconv"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// Test BluBeaver index page
func TestIndexPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode(BluBeaverURL)

	// Number of BD reviews listed on the index page
	numMovies := doc.Find("li a[href*='film' i][href$='htm' i]").Length()
	if numMovies < 5000 {
		log.Fatalln("Number of movie reviews seem really low", numMovies)
	}
}

// The "10" movie
// Normal movie review with large images available
func TestNormalMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("http://www.dvdbeaver.com/film3/blu-ray_reviews53/10_blu-ray.htm")

	// We should find 12 links to high-quality images
	numLargeImages := doc.Find("a[href*='large' i]:not([href*='subs' i])").Length()
	if numLargeImages != 12 {
		log.Fatalln("Number of large images is different than 12:", numLargeImages)
	}

	// since it's a standard movie page for a BD review
	// with high-quality images foundable as links,
	// we should not find and save other kind of images
	// such as inline images.
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

		// only consider images with decent resolution
		if heightInt >= 265 && widthInt >= 500 {
			numInlineImgs += 1
		}

		if numInlineImgs != 0 {
			log.Fatalln("Valid inline images should be zero, but found", numInlineImgs)
		}

	})
}

// 10,000 BC movie page
// No large image versions are available so
// we save the inlined images despite their average resolution
func TestMoviePageWithOnlyInlineImages(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("http://www.dvdbeaver.com/film2/DVDReviews38/10000_BC_blu-ray.htm")

	// No large links found, should return zero
	numLargeImages := doc.Find("a[href*='large' i]:not([href*='subs' i])").Length()
	if numLargeImages != 0 {
		log.Fatalln("Number of large images should be zero:", numLargeImages)
	}

	// Only inline images availabe, should have 12 of them
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

	if numInlineImgs != 12 {
		log.Fatalln("Valid inline images should be 12, but found", numInlineImgs)
	}

}

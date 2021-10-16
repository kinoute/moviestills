package websites

import (
	"log"
	"moviestills/utils"
	"testing"
)

// Test HighDefDiscNews index page
func TestHighDefDiscNewsIndexPage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode(HighDefDiscNewsURL)

	// Number of entries
	numMovies := doc.Find("div#mcTagMap ul.links a[href*=high]").Length()
	if numMovies < 200 {
		log.Fatalln("Number of movie reviews seem really low", numMovies)
	}
}

// Aliens
func TestHighDefDiscNewsNormalMoviePage(t *testing.T) {
	// Request the HTML page.
	doc := utils.GetHTMLCode("https://highdefdiscnews.com/2018/06/29/aliens-blu-ray-screenshots/")

	// We should find many links to high-quality images
	numLargeImages := doc.Find("div.gallery dl.gallery-item a[href*=high]").Length()
	if numLargeImages != 25 {
		log.Fatalln("Number of links to large images is different than 25:", numLargeImages)
	}
}

// Special function to remove some useless strings from links
func TestIsolateMovieTitle(t *testing.T) {
	cases := []struct {
		in       string
		expected string
	}{
		{"A Beautiful Day in the Neighborhood - Blu-ray Screenshots", "A Beautiful Day in the Neighborhood"},
		{"An American Werewolf in London [Limited Edition] - Blu-ray Screenshots", "An American Werewolf in London"},
		{"Re-Animator - Blu-ray Screenshots", "Re-Animator"},
		{"[REC] - Blu-ray Screenshots", "[REC]"},
		{"Another Movie", "Another Movie"},
		{"The Dead Zone - Blu-ray Review", "The Dead Zone"},
	}

	for _, c := range cases {
		got := isolateMovieTitle(c.in)
		if got != c.expected {
			t.Errorf("isolateMovieTitle(%q) == %q, expected %q", c.in, got, c.expected)
		}
	}
}

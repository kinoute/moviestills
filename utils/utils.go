package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// normalize movie's titles
// we need to handle extra multi spaces and weird accents
var normalizer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
var space = regexp.MustCompile(`\s+`)

func Normalize(str string) (string, error) {

	// get rid of accents
	s, _, err := transform.String(normalizer, str)
	if err != nil {
		return "", err
	}

	// get rid of trailing/leading accents and also multiple spaces
	s = strings.TrimSpace(s)
	s = space.ReplaceAllString(s, " ")
	return s, err
}

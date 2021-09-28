package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Normalize movies titles.
// We need to handle extra multi spaces and weird accents.
var normalizer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
var space = regexp.MustCompile(`\s+`)

// Remove GET parameters from URL.
// Sometimes we want to strip out forced dimensions
// or weird parameters that makes filenames weird.
var removeParams = regexp.MustCompile(`(\?.*)$`)

func RemoveURLParams(url string) string {
	return removeParams.ReplaceAllString(url, "")
}

func Normalize(str string) (string, error) {

	// Get rid of accents
	s, _, err := transform.String(normalizer, str)
	if err != nil {
		return "", err
	}

	// Get rid of trailing/leading accents and also multiple spaces
	s = strings.TrimSpace(s)
	s = space.ReplaceAllString(s, " ")
	return s, err
}

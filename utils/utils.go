package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Remove GET parameters from URL.
// Sometimes we want to strip out forced dimensions
// or weird parameters that makes filenames weird.
var removeParams = regexp.MustCompile(`(\?.*)$`)

func RemoveURLParams(url string) string {
	return removeParams.ReplaceAllString(url, "")
}

// Normalize movies titles.
// We need to handle extra multi spaces and weird accents.
var normalizer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
var space = regexp.MustCompile(`\s+`)

func Normalize(str string) (string, error) {
	// Get rid of accents
	s, _, err := transform.String(normalizer, str)
	if err != nil {
		return "", err
	}

	// Take care of disallowed characters for the creation
	// of folders and filenames on macOS/Linux/Windows.
	s = RemoveDisallowedChars(s)

	// Get rid of trailing/leading spaces and also multiple spaces
	s = strings.TrimSpace(s)
	s = space.ReplaceAllString(s, " ")
	return s, err
}

// Create (nested) folder if it doesn't exist yet
func CreateFolder(folder ...string) (string, error) {
	path := filepath.Join(folder...)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			return "", err
		}
	}
	return path, nil
}

// Remove disallowed characters from a string for
// the creation of folders and filenames on macOS/Linux/Windows.
// Taken from: https://github.com/iawia002/annie/blob/master/utils/utils.go
func RemoveDisallowedChars(name string) string {
	rep := strings.NewReplacer("\n", " ", "/", " ", "|", "-", ": ", "：", ":", "：", "'", "’")
	name = rep.Replace(name)
	if runtime.GOOS == "windows" {
		rep = strings.NewReplacer("\"", " ", "?", " ", "*", " ", "\\", " ", "<", " ", ">", " ")
		name = rep.Replace(name)
	}
	return name
}

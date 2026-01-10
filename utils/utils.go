package utils

import (
	"errors"
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

// removeParams is a regex to remove GET parameters from URL.
// Sometimes we want to strip out forced dimensions
// or weird parameters that makes filenames weird.
var removeParams = regexp.MustCompile(`(\?.*)$`)

// RemoveURLParams is used to remove URL Params
func RemoveURLParams(url string) string {
	return removeParams.ReplaceAllString(url, "")
}

// Normalize movies titles.
// We need to handle extra multi spaces and weird accents.
var normalizer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
var space = regexp.MustCompile(`\s+`)

// ErrEmptyResult is returned when normalization results in an empty string
var ErrEmptyResult = errors.New("empty result")

// ErrPathTraversal is returned when path traversal is detected
var ErrPathTraversal = errors.New("path traversal detected")

// Normalize a string. Remove anything weird from a string.
// In our case, movie titles. That way, creating folders
// or filenames won't be a problem.
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

	// Return an error if result is empty
	if s == "" {
		return "", ErrEmptyResult
	}

	// Check for path traversal attempts
	if strings.Contains(s, "..") {
		return "", ErrPathTraversal
	}

	return s, nil
}

// CreateFolder creates a (nested) folder if it doesn't exist yet
func CreateFolder(folder ...string) (string, error) {
	path := filepath.Join(folder...)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", err
	}
	return path, nil
}

// RemoveDisallowedChars removes disallowed characters from a string for
// the creation of folders and filenames on macOS/Linux/Windows.
// Taken from: https://github.com/iawia002/annie/blob/master/utils/utils.go
func RemoveDisallowedChars(name string) string {
	rep := strings.NewReplacer("\n", " ", "/", " ", "|", "-", ": ", "：", ":", "：", "'", "'")
	name = rep.Replace(name)
	if runtime.GOOS == "windows" {
		rep = strings.NewReplacer("\"", " ", "?", " ", "*", " ", "\\", " ", "<", " ", ">", " ")
		name = rep.Replace(name)
	}
	return name
}

// LimitLength handles overly long strings.
// Taken from: https://github.com/iawia002/annie/blob/master/utils/utils.go
func LimitLength(s string, length int) string {
	// 0 means unlimited
	if length == 0 {
		return s
	}

	const ellipses = "..."
	str := []rune(s)
	if len(str) > length {
		return string(str[:length-len(ellipses)]) + ellipses
	}
	return s
}

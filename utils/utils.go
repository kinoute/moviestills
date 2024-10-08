package utils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode"

	log "github.com/pterm/pterm"

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
		return "", errors.New("Empty result")
	}

	return s, err
}

// Create (nested) folder if it doesn't exist yet
func CreateFolder(folder ...string) (string, error) {
	path := filepath.Join(folder...)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", err
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

// LimitLength Handle overly long strings.
// Taken from: https://github.com/iawia002/annie/blob/master/utils/utils.go
func LimitLength(s string, length int) string {
	// 0 means unlimited
	if length == 0 {
		return s
	}

	const ELLIPSES = "..."
	str := []rune(s)
	if len(str) > length {
		return string(str[:length-len(ELLIPSES)]) + ELLIPSES
	}
	return s
}

// Generate a MD5 hash given a string
func MD5(fileName string) string {
	hasher := md5.New()
	hasher.Write([]byte(fileName))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Save movie image to the correct folder. Filenames
// can be hashed with MD5 if the option is set.
func SaveImage(moviePath, movieName, rawFileName string, body []byte, toHash bool) error {

	fileName := rawFileName
	extension := filepath.Ext(rawFileName)

	// Hash image filename with MD5 if asked to
	if toHash {
		fileName = MD5(rawFileName) + extension
	}

	outputImgPath := moviePath + "/" + fileName

	// Create nested folders, if needed
	if _, err := CreateFolder(moviePath); err != nil {
		return err
	}

	// Don't save again it we already downloaded it
	if _, err := os.Stat(outputImgPath); os.IsNotExist(err) {
		if err = os.WriteFile(outputImgPath, body, 0644); err != nil {
			return err
		}
	}

	// If we're here, image was successfully downloaded
	log.Success.Println("Saved image for", log.Blue(movieName), log.White(rawFileName))

	return nil
}

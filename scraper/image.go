package scraper

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
)

// MD5 generates an MD5 hash for the given string
func MD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// SaveImage saves a movie image to the correct folder.
// Filenames can be hashed with MD5 if the option is set.
func SaveImage(moviePath, movieName, rawFileName string, body []byte, toHash bool, log *Logger) error {
	fileName := rawFileName
	extension := filepath.Ext(rawFileName)

	// Hash image filename with MD5 if asked to
	if toHash {
		fileName = MD5(rawFileName) + extension
	}

	outputImgPath := filepath.Join(moviePath, fileName)

	// Create nested folders, if needed
	if err := os.MkdirAll(moviePath, os.ModePerm); err != nil {
		return err
	}

	// Don't save again if we already downloaded it
	if _, err := os.Stat(outputImgPath); os.IsNotExist(err) {
		if err = os.WriteFile(outputImgPath, body, 0644); err != nil {
			return err
		}
	}

	// If we're here, image was successfully downloaded
	log.Success("Saved image for", pterm.Blue(movieName), pterm.White(rawFileName))

	return nil
}

package main

import (
	"moviestills/config"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/pterm/pterm"
)

// printConfiguration prints configuration as a bullet list.
func printConfiguration(options *config.Options, websitesToScrape []string) {
	// Get fields and its values from the config struct
	values := reflect.ValueOf(*options)
	fields := values.Type()

	// Create bullet lists with configuration
	configuration := []pterm.BulletListItem{}

	// Add websites being scraped
	configuration = append(configuration, pterm.BulletListItem{
		Level:       0,
		Text:        pterm.Yellow("Websites") + ": " + pterm.Blue(strings.Join(websitesToScrape, ", ")),
		TextStyle:   pterm.NewStyle(pterm.FgBlue),
		BulletStyle: pterm.NewStyle(pterm.FgRed),
	})

	for i := 0; i < values.NumField(); i++ {
		fieldName := fields.Field(i).Name
		// Skip Website and All fields as we're showing them differently
		if fieldName == "Website" || fieldName == "All" {
			continue
		}
		configuration = append(configuration,
			pterm.BulletListItem{
				Level:       0,
				Text:        pterm.Yellow(fieldName) + ": " + pterm.Blue((values.Field(i).Interface())),
				TextStyle:   pterm.NewStyle(pterm.FgBlue),
				BulletStyle: pterm.NewStyle(pterm.FgRed),
			},
		)
	}

	// Print the configuration as a bullet list
	if err := pterm.DefaultBulletList.WithItems(configuration).Render(); err != nil {
		pterm.Error.Println("Could not print configuration", pterm.Red(err))
	}
}

// listAvailableScrapers prints list of available scrapers.
func listAvailableScrapers() {
	pterm.DefaultSection.Println("Scrapers available")

	// Get sorted list of scrapers
	scraperNames := make([]string, 0, len(sites))
	for name := range sites {
		scraperNames = append(scraperNames, name)
	}
	sort.Strings(scraperNames)

	// Create bullet lists with available scrapers
	availableScrapers := []pterm.BulletListItem{}
	for _, site := range scraperNames {
		availableScrapers = append(availableScrapers,
			pterm.BulletListItem{
				Level:       0,
				Text:        pterm.Yellow(site),
				TextStyle:   pterm.NewStyle(pterm.FgBlue),
				BulletStyle: pterm.NewStyle(pterm.FgRed),
			},
		)
	}

	// Print the available scrapers as a bullet list
	if err := pterm.DefaultBulletList.WithItems(availableScrapers).Render(); err != nil {
		pterm.Error.Println("Could not print available scrapers", pterm.Red(err))
		os.Exit(1)
	}

	// Show example of usage
	pterm.DefaultSection.Println("Usage")
	pterm.DefaultBasicText.Println("Single website:", pterm.Blue("--website"), pterm.White("blubeaver"))
	pterm.DefaultBasicText.Println("Multiple websites:", pterm.Blue("--website"), pterm.White("blubeaver"), pterm.Blue("--website"), pterm.White("film-grab"))
	pterm.DefaultBasicText.Println("All websites:", pterm.Blue("--all"))
	pterm.DefaultBasicText.Println("Sequential mode:", pterm.Blue("--all --sequential"))

	// Show how to contribute
	pterm.DefaultSection.Println("Contribution")
	pterm.DefaultBasicText.Println("See how you can add support for a new website:",
		pterm.White("https://github.com/kinoute/moviestills#contribute"),
	)
}

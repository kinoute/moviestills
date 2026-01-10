package main

import (
	"moviestills/config"
	"moviestills/utils"
	"os"
	"os/signal"
	"syscall"

	"github.com/pterm/pterm"
)

func clearScreen() {
	print("\033[H\033[2J")
}

func handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		pterm.Info.Println("Shutting down...")
		os.Exit(130)
	}()
}

func setupLogging(options *config.Options) {
	// Adjust the logging prefix
	pterm.Info = *pterm.Info.WithPrefix(pterm.Prefix{Text: " INFOS ", Style: pterm.Info.Prefix.Style})

	// Disable styles and colors based on options
	if options.NoStyle {
		pterm.DisableStyling()
	}
	if options.NoColors {
		pterm.DisableColor()
	}
}

func setupDirectories(options *config.Options) {
	// Create the cache directory
	if _, err := utils.CreateFolder(options.CacheDir); err != nil {
		pterm.Error.Println("The cache directory", pterm.White(options.CacheDir), "can't be created:", pterm.Red(err))
		os.Exit(1)
	}

	// Create the data directory
	if _, err := utils.CreateFolder(options.DataDir); err != nil {
		pterm.Error.Println("The data directory", pterm.White(options.DataDir), "can't be created:", pterm.Red(err))
		os.Exit(1)
	}
}

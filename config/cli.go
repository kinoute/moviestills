package config

// Description will be displayed on the CLI
// when a user makes use of the --help flag.
func (Options) Description() string {
	return "this program can scrap various websites to get high quality movie snapshots.\n"
}

// Epilogue will be displayed at the end
// of the "help" CLI command.
func (Options) Epilogue() string {
	return "For more information visit https://github.com/kinoute/moviestills"
}

// Version of the app can be displayed either
// when a user makes use of the --version flag, the
// --help flag or when an erroneous flag is passed.
func (Options) Version() string {
	return "moviestills " + VERSION + "\n"
}

package config

// Description will be displayed on the CLI
// when a user makes use of the --help flag.
func (Options) Description() string {
	return "this program can scrap various websites to get high quality movie snapshots.\n"
}

// Version of the app can be displayed either
// when a user makes use of the --version flag, the
// --help flag or when an erroneous flag is passed.
func (Options) Version() string {
	return "moviestills " + VERSION + "\n"
}

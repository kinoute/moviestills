package config

func (Options) Description() string {
	return "this program can scrap various websites to get high quality movie snapshots.\n"
}

func (Options) Version() string {
	return "moviestills " + VERSION + "\n"
}

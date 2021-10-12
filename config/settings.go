package config

// Options which can be set through the CLI or environment variables
type Options struct {
	Website  string `arg:"required,-w, --website,env:WEBSITE" help:"Website to scrap movie stills on"`
	Parallel int    `arg:"-p, --parallel,env:PARALLEL" help:"Limit the maximum parallelism" default:"2"`
	Async    bool   `arg:"-a, --async,env:ASYNC" help:"Enable asynchronus running jobs" default:"false"`
	CacheDir string `arg:"-c, --cache-dir,env:CACHE_DIR" help:"Where to cache scraped websites pages" default:"cache"`
	DataDir  string `arg:"-f, --data-dir,env:DATA_DIR" help:"Where to store movie snapshots" default:"data"`
	Debug    bool   `arg:"-d, --debug,env:DEBUG" help:"Enable Colly Debugger, our scraper" default:"false"`
	NoColors bool   `arg:"--no-colors,env:NO_COLORS" help:"Disable colors from output" default:"false"`
	Hash     bool   `arg:"--hash,env:HASH" help:"Hash image filenames with md5" default:"false"`
}

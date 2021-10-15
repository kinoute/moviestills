package config

import "time"

// Options which can be set through the CLI or environment variables
type Options struct {
	Website      string        `arg:"-w, --website,env:WEBSITE" help:"Website to scrap movie stills on"`
	ListScrapers bool          `arg:"-l, --list,env:LIST" help:"List all available scrapers implemented" default:"false"`
	Parallel     int           `arg:"-p, --parallel,env:PARALLEL" help:"Limit the maximum parallelism" default:"2"`
	RandomDelay  time.Duration `arg:"-r, --delay,env:RANDOM_DELAY" help:"Add some random delay between requests" default:"1s"`
	Async        bool          `arg:"-a, --async,env:ASYNC" help:"Enable asynchronus running jobs" default:"false"`
	CacheDir     string        `arg:"-c, --cache-dir,env:CACHE_DIR" help:"Where to cache scraped websites pages" default:"cache"`
	DataDir      string        `arg:"-f, --data-dir,env:DATA_DIR" help:"Where to store movie snapshots" default:"data"`
	Debug        bool          `arg:"-d, --debug,env:DEBUG" help:"Set Log Level to Debug to see everything" default:"false"`
	NoColors     bool          `arg:"--no-colors,env:NO_COLORS" help:"Disable colors from output" default:"false"`
	NoStyle      bool          `arg:"--no-style,env:NO_STYLE" help:"Disable styling and colors entirely from output" default:"false"`
	Hash         bool          `arg:"--hash,env:HASH" help:"Hash image filenames with md5" default:"false"`
}

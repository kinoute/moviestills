package scraper

import (
	"moviestills/config"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gocolly/colly/v2"
	"github.com/pterm/pterm"
)

// Stats tracks scraping progress
type Stats struct {
	Website          string
	MoviesFound      int64
	ImagesDownloaded int64
	ImagesFailed     int64
}

// Increment atomically increments a counter
func (s *Stats) IncrMovies() {
	atomic.AddInt64(&s.MoviesFound, 1)
}

func (s *Stats) IncrDownloaded() {
	atomic.AddInt64(&s.ImagesDownloaded, 1)
}

func (s *Stats) IncrFailed() {
	atomic.AddInt64(&s.ImagesFailed, 1)
}

// AggregatedStats holds stats from multiple scrapers
type AggregatedStats struct {
	mu      sync.Mutex
	PerSite []*Stats
	Total   Stats
}

// NewAggregatedStats creates a new aggregated stats tracker
func NewAggregatedStats() *AggregatedStats {
	return &AggregatedStats{
		PerSite: make([]*Stats, 0),
	}
}

// Add adds stats from a completed scraper
func (a *AggregatedStats) Add(s *Stats) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.PerSite = append(a.PerSite, s)
	a.Total.MoviesFound += s.MoviesFound
	a.Total.ImagesDownloaded += s.ImagesDownloaded
	a.Total.ImagesFailed += s.ImagesFailed
}

// Movie represents a movie being scraped
type Movie struct {
	Name string
	Year string
	URL  string
	Path string
}

// NewMovie creates a Movie with the proper path
func NewMovie(name, year, url, website string, options *config.Options) Movie {
	return Movie{
		Name: name,
		Year: year,
		URL:  url,
		Path: filepath.Join(options.DataDir, website, name),
	}
}

// ToContext stores movie data in a Colly context
func (m Movie) ToContext() *colly.Context {
	ctx := colly.NewContext()
	ctx.Put("movie_name", m.Name)
	ctx.Put("movie_path", m.Path)
	if m.Year != "" {
		ctx.Put("movie_year", m.Year)
	}
	return ctx
}

// MovieFromContext extracts movie data from a Colly context
func MovieFromContext(ctx *colly.Context) Movie {
	return Movie{
		Name: ctx.Get("movie_name"),
		Year: ctx.Get("movie_year"),
		Path: ctx.Get("movie_path"),
	}
}

// SiteConfig holds configuration for a specific website scraper
type SiteConfig struct {
	Name           string
	IndexURL       string
	AllowedDomains []string
	DetectCharset  bool
}

// SetupIndexScraper configures the main index scraper with common settings
func SetupIndexScraper(c *colly.Collector, cfg SiteConfig, log *Logger) {
	c.AllowedDomains = cfg.AllowedDomains
	c.AllowURLRevisit = true

	if cfg.DetectCharset {
		c.DetectCharset = true
	}

	// Common error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Error(r.Request.URL, "\t", pterm.White(r.StatusCode), "\nError:", pterm.Red(err))
	})

	// Common request handler
	c.OnRequest(func(r *colly.Request) {
		log.Debug("visiting index page", pterm.White(r.URL.String()))
	})
}

// SetupMovieScraper configures the movie page scraper with common settings
func SetupMovieScraper(c *colly.Collector, log *Logger) *colly.Collector {
	movieScraper := c.Clone()
	movieScraper.AllowURLRevisit = false

	movieScraper.OnRequest(func(r *colly.Request) {
		log.Debug("visiting", pterm.White(r.URL.String()))
	})

	return movieScraper
}

// SetupImageResponseHandler sets up the common image response handler
func SetupImageResponseHandler(c *colly.Collector, options *config.Options, stats *Stats, log *Logger) {
	c.OnResponse(func(r *colly.Response) {
		// Ignore anything that is not an image
		if !strings.Contains(r.Headers.Get("Content-Type"), "image") {
			return
		}

		movie := MovieFromContext(r.Ctx)

		if err := SaveImage(movie.Path, movie.Name, r.FileName(), r.Body, options.Hash, log); err != nil {
			log.Error("Can't save image", pterm.White(r.FileName()), pterm.Red(err))
			if stats != nil {
				stats.IncrFailed()
			}
			return
		}

		if stats != nil {
			stats.IncrDownloaded()
		}
	})
}

// VisitAndWait visits the index URL and waits for completion
func VisitAndWait(indexScraper *colly.Collector, movieScraper *colly.Collector, url string, log *Logger) {
	if err := indexScraper.Visit(url); err != nil {
		log.Error("Can't visit index page", pterm.White(url), ":", pterm.Red(err))
	}

	indexScraper.Wait()
	if movieScraper != nil {
		movieScraper.Wait()
	}
}

// PrintSummary prints the final scraping statistics for a single site
func PrintSummary(stats *Stats) {
	pterm.DefaultSection.Println("Summary")
	printStatsItems(stats)
}

// PrintAggregatedSummary prints stats for all scraped sites
func PrintAggregatedSummary(agg *AggregatedStats) {
	pterm.DefaultSection.Println("Summary")

	// Print per-site stats if multiple sites were scraped
	if len(agg.PerSite) > 1 {
		for _, s := range agg.PerSite {
			pterm.DefaultSection.WithLevel(2).Println(s.Website)
			printStatsItems(s)
		}
		pterm.DefaultSection.WithLevel(2).Println("Total")
	}

	printStatsItems(&agg.Total)
}

func printStatsItems(stats *Stats) {
	items := []pterm.BulletListItem{
		{
			Level:       0,
			Text:        pterm.Sprintf("Movies found: %s", pterm.White(stats.MoviesFound)),
			TextStyle:   pterm.NewStyle(pterm.FgDefault),
			BulletStyle: pterm.NewStyle(pterm.FgGreen),
		},
		{
			Level:       0,
			Text:        pterm.Sprintf("Images downloaded: %s", pterm.White(stats.ImagesDownloaded)),
			TextStyle:   pterm.NewStyle(pterm.FgDefault),
			BulletStyle: pterm.NewStyle(pterm.FgGreen),
		},
		{
			Level:       0,
			Text:        pterm.Sprintf("Images failed: %s", pterm.White(stats.ImagesFailed)),
			TextStyle:   pterm.NewStyle(pterm.FgDefault),
			BulletStyle: pterm.NewStyle(pterm.FgRed),
		},
	}

	if err := pterm.DefaultBulletList.WithItems(items).Render(); err != nil {
		pterm.Error.Println("Could not print summary", pterm.Red(err))
	}
}

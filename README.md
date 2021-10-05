# Movie Stills

A Go CLI application to scrap various websites in order to get high-quality movie snapshots.

## Installation

There are various ways to install or use the application:

### Binaries

Download the latest binary from the [releases](https://github.com/kinoute/moviestills/releases) for your OS.

Then you can simply execute the binary like this (on Linux):

```bash
./moviestills --help
```

See [Usage](#Usage) to check what settings you can pass through CLI arguments.

### Docker image

Docker comes to the rescue, providing an easy way how to run `moviestills` on most platforms:

```bash
docker run 
    --name moviestills \
    --rm hivacruz/moviestills:latest \
    --website movie-screencaps \
    --async
```

You can also use environment instead of CLI arguments:

```bash
docker run \
    --name moviestills
    -e WEBSITE=blubeaver \
    -e ASYNC=true \
    --rm hivacruz/moviestills:latest
```

See [Usage](#Usage) to check what settings you can pass as environment variables.

### Clone the repo

If you already have Go installed, you can also simply clone the repo and run the application from the folder:

```bash
# clone the repo
git clone https://github.com/kinoute/moviestills

# go inside the project
cd moviestills

# build the binary
go build

# use the app (example on Linux)
./moviestills --help

# you can also run the app without compiling
go run . --help
```

## Usage

```bash
Usage: moviestills --website WEBSITE [--parallel PARALLEL] [--async] [--cache-dir CACHE-DIR] [--debug]

Options:
  --website WEBSITE, -w WEBSITE
                         Website to scrap movie stills on [env: WEBSITE]
  --parallel PARALLEL, -p PARALLEL
                         Limit the maximum parallelism [default: 2, env: PARALLEL]
  --async, -a            Enable asynchronus running jobs [env: ASYNC]
  --cache-dir CACHE-DIR, -c CACHE-DIR
                         Where to cache scraped websites pages [default: cache, env: CACHE_DIR]
  --debug, -d            Enable debugging on Colly, our scraper library [env: DEBUG]
  --help, -h             display this help and leave
  --version              display version and leave
```

**Note**: CLI arguments will always override environment variables. Therefore, if you set `WEBSITE` as an environment variable and also use `—website` as a CLI argument, only the latter will be passed to the app.

For boolean arguments such as `--async` or `--debug`, their equivalent as environment variables is, for example, `ASYNC=false` or `ASYNC=true`.

### Cache

By default, every scraped page will be cached in a `cache` folder. You can change this through the options as listed above. This is an important folder as it stores everything and avoid requesting again some websites pages when there is no need to.

### Data

By default, each scraped website will have its own subfolder in `data`. It will contain the movie snapshots for each movie found on the website.

Example:

```shell
data/blubeaver
├── 12\ Angry\ Men
│   ├── film3_blu_ray_reviews55_12_angry_men_blu_ray_large_large_12_angry_men_blu_ray_1.jpg
│   ├── film3_blu_ray_reviews55_12_angry_men_blu_ray_large_large_12_angry_men_blu_ray_1x.jpg
│   ├── film3_blu_ray_reviews55_12_angry_men_blu_ray_large_large_12_angry_men_blu_ray_2.jpg
```

# Movie Stills

[![CI](https://github.com/kinoute/moviestills/actions/workflows/ci.yml/badge.svg)](https://github.com/kinoute/moviestills/actions/workflows/ci.yml)
[![Go Report](https://goreportcard.com/badge/github.com/kinoute/moviestills)](https://goreportcard.com/report/github.com/kinoute/moviestills)

A Go CLI application to scrap various websites in order to get high-quality movie snapshots. See the list of the [Supported Websites](#supported-websites).

## Installation

There are various ways to install or use the application:

### Binaries

Download the latest binary from the [releases](https://github.com/kinoute/moviestills/releases) page for your OS. Then you can simply execute the binary like this:

```bash
# example on linux
./moviestills --help

# You can also use environment variables 
# instead of CLI arguments
WEBSITE=blubeaver ./moviestills
```

See [Usage](#Usage) to check what settings you can pass through CLI arguments or environment variables.

### Docker images

Docker comes to the rescue, providing an easy way how to run `moviestills` on most platforms.

#### GitHub Registry

```bash
docker run \
    --name moviestills \
    -v "$PWD/cache:/cache" \ # cache scraped websites pages
    -v "$PWD/data:/data" \ # store movie snapshots
    --rm ghcr.io/kinoute/moviestills:latest \
    --website movie-screencaps \ # scrap a specific website
    --async # enable asynchronous jobs
```

#### Docker Hub

```bash
docker run \
    --name moviestills \
    -v "$PWD/cache:/cache" \ # cached scraped websites
    -v "$PWD/data:/data" \ # store movie snapshots
    -e WEBSITE=blubeaver \
    -e ASYNC=true \
    --rm hivacruz/moviestills:latest
```

As you an see, you can also use environment instead of CLI arguments.

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

By default, every scraped page will be cached in the `cache` folder. You can change the name or path to the folder  through the options as listed above, with `—cache-dir` or the `CACHE_DIR` environment variable. This is an important folder as it stores everything that was scraped.

It avoids requesting again some websites pages when there is no need to. It is a nice thing as we don't want to flood these websites with thousands of useless requests.

In case you're using our Docker images to run `moviestills`, don't forget to change the volume path to the new internal cache folder, if you set up a custom cache folder.

### Data

By default, each scraped website will have its own subfolder in the  `data` folder. Inside, every movie will have its own folder with the scraped movie snapshots found on the website.

Example:

```shell
data # where to store movie snapshots
├── blubeaver # website's name
│   ├── 12\ Angry\ Men # movie's title
│   │   ├── film3_blu_ray_reviews55_12_angry_men_blu_ray_large_large_12_angry_men_blu_ray_1.jpg
│   │   ├── film3_blu_ray_reviews55_12_angry_men_blu_ray_large_large_12_angry_men_blu_ray_1x.jpg
│   │   ├── film3_blu_ray_reviews55_12_angry_men_blu_ray_large_large_12_angry_men_blu_ray_2.jpg
```

## Supported Websites

As today, scrapers were implemented for the following websites in `moviestills`:

| Website                                        | Simplified Name [1]() | Description                                                  | Number of movies |
| ---------------------------------------------- | --------------------- | ------------------------------------------------------------ | ---------------- |
| [BluBeaver](http://blubeaver.ca)               | blubeaver             | Extension of [DVDBeaver](http://dvdbeaver.com), this time dedicated to Blu-Ray reviews only. Reviews are great to check the quality of BD releases with lot of technical details. Only snapshots on "free" access are scraped. | ~3567            |
| [BlusScreens](https://www.bluscreens.net)      | bluscreens            | Website with high resolution screen captures taken directly from different Blu-ray releases by [Blusscreens](https://twitter.com/Bluscreens). | ~452             |
| [DVDBeaver](http://dvdbeaver.com)              | dvdbeaver             | ***Not recommended***. [2]() A massive list of DVD/BD reviews with a lot of movie snapshots. It includes the BD reviews available in [BluBeaver](http://blubeaver.ca). It is advised to use the latter instead. | ~9975            |
| [EvanERichards](https://www.evanerichards.com) | evanerichards         | A short but interesting list of movies with a lot of snapshots for each. Also includes some TV Series but they are ignored by the scraper. | ~245             |
| [Film-Grab](https://film-grab.com)             | film-grab             | A great list of movies with a few snapshots for each. Snapshots were cherry-picked and show nice cinematography. | ~2829            |

[1](). The name column displays the website's name to use with `moviestills` to start the scraping job. Eg `—website blubeaver` for the BluBeaver.ca website.

[2](). While DVDBeaver provides a lot of great movie snapshots from DVD reviews, it is harder to filter correctly the movie snapshots on the reviews pages. Expect a lot of false positives (DVD covers, banners etc).

## Development

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

## 


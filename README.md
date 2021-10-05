# Movie Stills

[![CI](https://github.com/kinoute/moviestills/actions/workflows/ci.yml/badge.svg)](https://github.com/kinoute/moviestills/actions/workflows/ci.yml)
[![Go Report](https://goreportcard.com/badge/github.com/kinoute/moviestills)](https://goreportcard.com/report/github.com/kinoute/moviestills)

A Go CLI application to scrap various websites in order to get high-quality movie snapshots. See the list of [Supported Websites](#supported-websites).

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
    --volume "$PWD/cache:/cache" \ # cache scraped websites pages
    --volume "$PWD/data:/data" \ # store movie snapshots
    --rm ghcr.io/kinoute/moviestills:latest \
    --website movie-screencaps \ # scrap a specific website
    --async # enable asynchronous jobs
```

#### Docker Hub

```bash
docker run \
    --name moviestills \
    --volume "$PWD/cache:/cache" \ # cached scraped websites
    --volume "$PWD/data:/data" \ # store movie snapshots
    --env WEBSITE=blubeaver \
    --env ASYNC=true \
    --rm hivacruz/moviestills:latest
```

As you can see, you can also use environment variables instead of CLI arguments.

## Usage

```bash
Usage: moviestills --website WEBSITE 
									 [--parallel PARALLEL] [--async] [--cache-dir CACHE-DIR] [--debug]

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

In case you are using our Docker images to run `moviestills`, don't forget to change the volume path to the new internal cache folder, if you set up a custom cache folder.

### Data

By default, each scraped website will have its own subfolder in the `data` folder. Inside, every movie will have its own folder with the scraped movie snapshots found on the website.

Example:

```shell
data # where to store movie snapshots
├── blubeaver # website's name
│   ├── 12\ Angry\ Men # movie's title
│   │   ├── film3_blu_ray_reviews55_12_angry_men_blu_ray_large_large_12_angry_men_blu_ray_1.jpg
│   │   ├── film3_blu_ray_reviews55_12_angry_men_blu_ray_large_large_12_angry_men_blu_ray_2.jpg
│   │   ├── film3_blu_ray_reviews55_12_angry_men_blu_ray_large_large_12_angry_men_blu_ray_3.jpg
```

## Supported Websites

As today, scrapers were implemented for the following websites in `moviestills`:

| Website                                        | Simplified Name [<sup>1</sup>]() | Description                                                  | Number of movies [<sup>3</sup>]() |
| ---------------------------------------------- | ------------------------------- | ------------------------------------------------------------ | ---------------- |
| [BluBeaver](http://blubeaver.ca)               | blubeaver                       | Extension of [DVDBeaver](http://dvdbeaver.com), this time dedicated to Blu-Ray reviews only. Reviews are great to check the quality of BD releases with lot of technical details. Only snapshots on "free" access are scraped. | ~3567            |
| [BlusScreens](https://www.bluscreens.net)      | bluscreens                      | Website with high resolution screen captures taken directly from different Blu-ray releases by [Blusscreens](https://twitter.com/Bluscreens). | ~452             |
| [DVDBeaver](http://dvdbeaver.com)              | dvdbeaver                       | ***Not recommended***. [<sup>2</sup>]() A massive list of DVD/BD reviews with a lot of movie snapshots. It includes the BD reviews available in [BluBeaver](http://blubeaver.ca). It is advised to use the latter instead. | ~9975            |
| [EvanERichards](https://www.evanerichards.com) | evanerichards                   | A short but interesting list of movies with a lot of snapshots for each. Also includes some TV Series but they are ignored by the scraper. | ~245             |
| [Film-Grab](https://film-grab.com)             | film-grab                       | A great list of movies with a few snapshots for each. Snapshots were cherry-picked and show nice cinematography. | ~2829            |
| [HighDefDiscNews](https://highdefdiscnews.com) | highdefdiscnews                 | A few hundreds movies featured with high-quality snapshots (png, lossless) in native resolution. | ~209             |
| [Movie-Screencaps](https://movie-screencaps.com) | movie-screencaps | Website with DVD, BD, and 4K BD movie snapshots. Since hundreds of snapshots are available for each movie (one per second or so), we only take some of them per paginated page. | ~715 |
| [ScreenMusings](https://screenmusings.org) | screenmusings | A small list of movies but with nice cherry-picked snapshots. | ~260 |
| [StillsFrmFilms](https://stillsfrmfilms.wordpress.com) | stillsfrmfilms | A very small list of movies but, again, the snapshots were nicely chosen and depict perfectly the atmosphere of each movie. | ~63 |

[<sup>1</sup>]() : The name column displays the website's name to use with `moviestills` to start the scraping job. Eg `—website blubeaver` for the BluBeaver.ca website.

[<sup>2</sup>]() : While DVDBeaver provides a lot of great movie snapshots from DVD reviews, it is harder to filter correctly the movie snapshots on the reviews pages. Expect a lot of false positives (DVD covers, banners etc).

[<sup>3</sup>]() : Approximate number of movies calculated on October 5th, 2021. 

**Contribute:** If you want to add a new website to the scraper, please read how to set up a [development workflow](#development) and [how to contribute](#contribute).

## Development

### Clone the repo

If you already have Go installed, you can simply clone the repo and run the application from the folder:

```bash
# clone the repo
git clone https://github.com/kinoute/moviestills

# go inside the project
cd moviestills
```

Then:

```bash
# download the dependencies
go mod download

# you can run the app without compiling
go run . --help

# optional: compile the binary
go build -v

# use the compiled app (example on Linux)
./moviestills --help
```

### Docker

If you don't have Go installed, you can build a Docker Image to start developing in a Go environment. To do that, clone the repository and inside the folder:

```bash
# build development image
docker --tag moviestills-dev . --target base
```

Then start and go inside the container:

```bash
# start the development container
# and go inside it
docker run \
	--name moviestills-dev \
	--volume "$PWD:/app" \
	--interactive \
	--tty \
	--rm moviestills-dev
```

A volume is created to report any change you make to the code inside the container.

To run your code, you can use `go run .` inside the container, test it, build it etc. Do note that these commands might be slow at first.

## Contribute

This is my first project in Golang. Therefore, pull requests, suggestions or bug reports are appreciated. A major refactoring is not excluded while I still learn the language.

### Add a new website

You can contribute to this scraper by adding a new website which provides high-quality movie snapshots. To do that, there are four steps:

1. Create a new file in the `websites` folder with the *simplified* name of the website (eg. `yahoo.go`). Check how other websites were implemented and scraped with the [Colly](https://github.com/gocolly/colly) library. You need to define a main URL as a constant (the first URL that will get visited) and a main function such as `yahooScraper()` where your Colly logic will do the work.
2. Once you created a scraper for a website, you need to add this new scraper in the available options of the app in `main.go`. Edit the `sites` map by adding the *simplified* name of the website as a key and add the scraper's function as a value (eg. `websites.yahooScraper`).
3. Create a unit test for the website, eg `yahoo_test.go`. For that test, we are not going to test with Colly but only with [GoQuery](https://github.com/PuerkitoBio/goquery), a library that makes HTML/CSS parsing easy. We just want to make sure the CSS selectors we use in our scraper are still up-to-date and are still filtering correctly the data we are looking for.
4. Edit the [Supported Websites](#supported-websites) table in the `README` file and write detailed informations about the website you added – please make sure websites are sorted alphabetically in the table.

## Support

Most of the websites we are scraping are owned by individuals who just want to share nice movie snapshots. **Please don't abuse these websites and limit your scraping activity to not arm them in any way.**

You can support some of the webmasters behind these nice galleries:

* The owner of DVDBeaver/BluBeaver has a Patreon page where you can support him and also get access to a lot more movie snapshots in great quality. Here is the link: https://www.patreon.com/dvdbeaver ;
* The man behind Film-Grab also has a Patreon page where you can send him some love: https://www.patreon.com/filmgrab.

I couldn't find any support page for the other websites but you can support some of them by using their affiliated links for example.

Just be gentle while scraping: some are hosting the images on their own servers!


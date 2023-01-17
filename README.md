# Iemscrapper - 3-in-1 application (daemon, web scrapper, http api)
App works as a daemon, periodically scrapping data from popular earphone store and one (for now) earphone ranking site. The data is then combined, archived in MongoDB database with Redis cache. Data is exposed via simple http api (for latest, and all scrap sessions).

Application lacks the front-end interface for now, but I plan on adding that later.

The app is probably extremely overengineered, but was a nice intro to working with Redis & MongoDB on a real example.

## Features:
- Web scraper setup to gather data from earphone store and earphone ranking sites
- Automatic periodic scraping using go-cron library
- Scrapped data is served through two endpoints ('/latest' for latest scraping session from Redis cache and '/all' for all scrapped data directly from MongoDB database)
- Ability to pass your own service URIs through command line arg

## Running:
After starting MongoDB & Redis services:
`go run . -db_uri="<mongodb_uri>" -cache_uri="<redis_uri>"`

For now the app is hardcoded to scrap every 12h, but eventually that could be passed as args too.

## Todo:
- Add more ranking sites (and potentially shops)
- Fix temporary context.TODO()
- Rank matching function refactoring (more efficient algorithm and wrapped in goroutine)

## Stack:
- Go 1.19+ ([GoColly - web scrapper](https://github.com/gocolly/colly), [gocron - scheduler](https://github.com/go-co-op/gocron)]
- MongoDB 6.0+
- Redis 7+

## Credits:
- none

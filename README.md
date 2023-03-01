# SC Mixpanel to Posthog Data Migrator

A tool for easily migrating data from [Mixpanel](https://mixpanel.com/) to [Posthog](https://posthog.com) (self-hosted or cloud)

## Disclaimer

This is **NOT an official tool**. We are not affiliated with Posthog, Mixpanel, or any other project.

However, if you are looking to migrate from Mixpanel to Posthog like we were - we hope you find this tool useful.

## Setup

### Mixpanel

You will need the following from Mixpanel:

- Service account username and password (with owner privileges)
- The project ID (found in settings -> overview)

You will be prompted in CLI to input these, or you can set the following env variables:

```
MIXPANEL_USERNAME=
MIXPANEL_PASSWORD=
MIXPANEL_PROJECT_ID=
# Optional override, defaults to https://data.mixpanel.com/api/2.0
MIXPANEL_API_URL
```

You can also put these in `.env` for convenience.

### Posthog

You will need the following from Posthog:

- Project API key
- Personal API key
- Endpoint URL

You will be prompted in CLI for these, but can also set them in the environment:

```
POSTHOG_PROJECT_KEY=
POSTHOG_API_KEY=
POSTHOG_ENDPOINT=
```

## **WARNING** Do not use this without reading this first.

The mixpanel export API has no pagination, the CLI will prompt you for a date range (required by Mixpanel)

If you have a very large data set, **do not try to get it all at once**

Mixpanel could rate limit you, your system could run out of memory and crash.

It's recommended to do smaller chunks at a time (dates are inclusive, so from_date=2023-03-01 and to_date=2023-03-01 will import 1 days worth of data)

# Usage

Download the latest [Release](https://github.com/stablecog/sc-mp-to-ph/releases) for your system.

Simply run it with, `./migrator` or `./migrator.exe` if using windows.

## Import Users

The mixpanel Web UI allows exporting users as csv format. Select all columns, all users, get a .csv file.

Run `./migrator -users-csv-file /path/to/users-export.csv` to load all users into mixpanel

## Usage (Import Events)

Just run without any parameters.

## Check us out

[Stablecog](https://stablecog.com/)

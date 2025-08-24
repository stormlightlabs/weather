# Stormlight Weather API

A comprehensive weather API service that aggregates data from multiple official weather services including NOAA NWS and international sources,
with built-in geocoding capabilities using the US Census Geocoder.

## Architecture

## Toolchain

Structured logging with charmbracelet/log

Swaggo for OpenAPI docs

Github OAuth

Postgres & KeyDB (Request Caching)

## Data Sources

| Source                 | Description                                            | Coverage                | Notes                                                                                                                       |
| :--------------------- | :----------------------------------------------------- | :---------------------- | :-------------------------------------------------------------------------------------------------------------------------- |
| **NOAA (NWS)**         | US National Weather Service (official US weather data) | USA only                | For US forecasts, alerts, radar. [API](https://www.weather.gov/documentation/services-web-api)                              |
| **Met.no (Yr API)**    | Norwegian Meteorological Institute, "met.no"           | Global                  | Very popular (used by open-meteo, car apps, etc). [API](https://api.met.no/weatherapi/documentation) - attribution required |
| **DWD**                | Deutscher Wetterdienst (Germany)                       | Germany + global models | [Open Data](https://opendata.dwd.de/) FTP downloads; no JSON API directly                                                   |
| **ECMWF**              | European Centre for Medium-Range Weather Forecasts     | Global (weather models) | Requires some access setup for APIs                                                                                         |
| **Environment Canada** | Canadian Meteorological Data                           | Canada                  | [Weather API Guide](https://weather.gc.ca/mainmenu/about_envcan_e.html) (RSS/XML mostly)                                    |
| **MeteoSwiss**         | Swiss Meteorological Data                              | Switzerland             | [Data](https://www.meteoswiss.admin.ch/) - mostly local                                                                     |
| **Copernicus (EU)**    | Satellite data (climate, atmospheric data)             | Global                  | [Open Access Hub](https://scihub.copernicus.eu/)                                                                            |

---

| Source                        | Description                              | Coverage                | Notes                                                                                                                          |
| :---------------------------- | :--------------------------------------- | :---------------------- | :----------------------------------------------------------------------------------------------------------------------------- |
| **Nominatim (OpenStreetMap)** | Open-source geocoding engine             | Global                  | Can self-host, no cost. Respect rate limits if public instance. [Docs](https://nominatim.org/release-docs/develop/api/Search/) |
| **US Census Geocoder**        | US address geocoding                     | USA only                | [API](https://geocoding.geo.census.gov/geocoder/)                                                                              |
| **Geonames**                  | Open geodata (cities, places, elevation) | Global                  | Free with attribution. [API](http://www.geonames.org/export/web-services.html)                                                 |
| **Natural Earth Data**        | Static global country/city boundaries    | Global                  | [Data](https://www.naturalearthdata.com/) (not an API, but for static data)                                                    |
| **GADM**                      | Administrative boundaries                | Global                  | Shapefiles for countries, provinces, etc.                                                                                      |
| **OpenAddresses**             | Open-source address points               | Global (some countries) | Raw dumps, no API. Self-hosting needed                                                                                         |

## Encryption

The secrets package provides secure configuration management using AES-256-GCM encryption with scrypt key derivation for sensitive environment variables (see [`env.local`](env.local)).
Keys can be sourced from CLI arguments, environment variables, or interactive prompts, with built-in validation requiring 12+ characters, mixed case, digits, and forbidden pattern detection. Values are encrypted in a salt:nonce:ciphertext format, allowing seamless handling where non-encrypted values pass through unchanged while encrypted values are automatically decrypted when accessed.

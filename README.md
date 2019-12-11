# gramarr
A [Radarr](https://github.com/Radarr/Radarr)/[Sonarr](https://github.com/Sonarr/Sonarr) Telegram Bot featuring user authentication/level access.

## Features

### Sonarr

- Search for TV Shows by name. - does not work for now
- Pick which seasons you want to download.
- Choose which quality and language profile you want to download.

### Radarr

- Search for Movies by name.
- Choose which quality profile you want to download.

## Requirements

### If running from source

- Go

## Configuration

- Copy the `config.json.template` file to `config.json` and set-up your configuration;
- Put the `config.json` alongside with the binary downloaded from [releases](https://github.com/alcmoraes/gramarr/releases);

## Running it

### From source

```bash
$ go get github.com/memodota/gramarr
$ cd $GOPATH/src/github.com/memodota/gramarr
$ go run .
```

### From release

Just [download](https://github.com/memodota/gramarr/releases/latest) the respective binary for your System.

*Obs: Don't forget to put the `config.json` in the same folder as the binary file.*

### In docker

- copy file `config.json.template` to `config.json` to "yourwayonhost"
- docker pull memodota/gramarrru:latest
- run in docker   -   docker run -d --name=sonarr-radarr-telegram-bot --restart=always -v yourwayonhost:/config memodota/gramarrru

## TODO

- Fully translate to Russian
- Make language selectable in config file
- Fix Adding TV Shows

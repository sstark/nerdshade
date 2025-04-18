## nerdshade

> [!WARNING]
> Early development stage

Calculates outside brightness based on location and local time.

Brightness value transitions smoothly for an hour during sunrise and sunset.

Brightness is translated into a color temperature and handed over to [hyprsunset](https://github.com/hyprwm/hyprsunset).

Actual calculation of sunrise/sunset times is done by the [go-sunrise package](https://github.com/nathan-osman/go-sunrise).

Can be run in one-shot mode (default) or in a loop.

## Usage

```
$ ./nerdshade -h
Usage of ./nerdshade:
  -V    Show program version
  -debug
        Print debug info
  -gammaDay int
        Day gamma (default 100)
  -gammaNight int
        Night gamma (default 90)
  -latitude float
        Your location latitude (default 48.516)
  -longitude float
        Your location longitude (default 9.12)
  -loop
        Run nerdshade continuously
  -tempDay int
        Day color temperature (default 6500)
  -tempNight int
        Night color temperature (default 4000)
```

## Installation

It's still in development, so installation is a bit rough:

Make sure you have hyprsunset running.

- Clone repository
- `TZ=CET go test`
- `go build`
- `mkdir -p ~/bin`        # adjust
- `cp nerdshade ~/bin`    # adjust
- `hyprctl keyword exec hyprsunset`    # if not yet running
- `hyprctl keyword exec "nerdshade -loop"`

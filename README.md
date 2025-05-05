## nerdshade 🕶

Calculates outside brightness based on location and local time OR based on
fixed schedule.

Brightness is translated into a color temperature and gamma value and handed
over to [hyprsunset](https://github.com/hyprwm/hyprsunset).

Color temperature and gamma values transition smoothly for an hour
(configurable) during sunrise and sunset (or wakupe/bedtime respectively).

Actual calculation of sunrise/sunset times is done by the [go-sunrise package](https://github.com/nathan-osman/go-sunrise).

Can be run in one-shot mode (default) or in a loop.

## Usage

```
$ ./nerdshade -h
Usage of ./nerdshade:
  -V    Show program version
  -debug
        Print debug info
  -fixedBedtime string
        Bedtime time in 24-hour format, e. g. "22:30" (overrides location)
  -fixedWakeup string
        Wakeup time in 24-hour format, e. g. "6:00" (overrides location)
  -gammaDay int
        Day gamma (default 100)
  -gammaNight int
        Night gamma (default 90)
  -hyperctl string
        Path to hyperctl program (default "hyprctl")
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
  -transitionDuration duration
        Duration of transition, e. g. "45m" or "1h10m" (default 1h0m0s)
```

## Installation (Arch / AUR)

For example with `yay`:

```sh
yay -S nerdshade
```

## Installation (Other)

Make sure you have hyprsunset running.

Download the latest binary from releases, place it somwhere in `$PATH` and start it. Example:

- (download)
- `cp nerdshade ~/.local/bin/nerdshade`
- `chmod +x ~/.local/bin/nerdshade`
- `hyprctl keyword exec hyprsunset`    # if not yet running
- `hyprctl keyword exec "nerdshade -loop"` # adjust

## Building

- Clone repository
- `make test` # optional
- `make`

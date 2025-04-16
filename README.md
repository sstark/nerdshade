## nerdshade

> [!WARNING]
> This is early work in progress and not yet complete.

Calculates outside brightness based on location and local time.

Brightness value transitions smoothly for an hour during sunrise and sunset.

Brightness is translated into a color temperature and handed over to [hyprsunset](https://github.com/hyprwm/hyprsunset).

Actual calculation of sunrise/sunset times is done by the [go-sunrise package](https://github.com/nathan-osman/go-sunrise).

## Usage

```
./nerdshade -h
Usage of ./nerdshade:
  -debug
        Print debug info
  -latitude float
        Your location latitude (default 48.516)
  -longitude float
        Your location longitude (default 9.12)
  -max int
        Maximum color temperature (default 6500)
  -min int
        Minimum color temperature (default 4000)
```

## nerdshade 🕶

Calculates outside brightness based on location and local time OR based on
fixed schedule.

Brightness is translated into a color temperature and gamma value and handed
over to [hyprsunset](https://github.com/hyprwm/hyprsunset).

Color temperature and gamma values transition smoothly for an hour
(configurable) during sunrise and sunset (or wakupe/bedtime respectively).

Actual calculation of sunrise/sunset times is done by the [go-sunrise package](https://github.com/nathan-osman/go-sunrise).

Can be run in one-shot mode (default) or in a loop.

If nerdshade runs in loop mode and the `acpi_listen` program could be found,
nerdshade will immediately do an update when the laptop lid is opened. (This
is still a bit experimental, feedback is welcome)

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
  -nologtime
        Omit time from log output (e. g. when running as systemd unit)
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

## Installation using Nix Home Manager Flake

Add

```
nerdshade.url = "github:sstark/nerdshade";
```

as an input in your `flake.nix` and import the home-manager module, e. g.:

```nix
homeConfigurations."you@host" = home-manager.lib.homeManagerConfiguration {
  pkgs = nixpkgs.legacyPackages.x86_64-linux;
  modules = [
    nerdshade.homeModules.default
    ./home.nix
  ];
};
```

Then enable and configure it in your home-manager module:

```nix
services.nerdshade = {
  enable = true;
  latitude = 50.534;
  longitude = 8.6;
  tempNight = 4200;
};
```

This installs the package, writes the settings to
`$XDG_CONFIG_HOME/nerdshade.yaml` and runs nerdshade in loop mode via a
systemd user service. Set `services.nerdshade.systemd.enable = false` if you
do not want the service, or use `services.nerdshade.systemd.extraArgs` to
pass additional command-line arguments. All command line options are available
as module options. See `hm-module.nix` for the full list.

### Try without Installation

To try it out without installation you can use:

```
nix run github:sstark/nerdshade -- --help
```

You can easily try the dev version this way:

```
nix run github:sstark/nerdshade/dev -- --help    
```

## Manual Installation

Make sure you have hyprsunset running.

Download the latest binary from releases, place it somwhere in `$PATH` and start it. Example:

- (download)
- `cp nerdshade ~/.local/bin/nerdshade`
- `chmod +x ~/.local/bin/nerdshade`
- `hyprctl keyword exec hyprsunset`    # if not yet running
- `hyprctl keyword exec "nerdshade -loop"` # adjust


## Configuration File

Parameters may be defined via a configuration file in `$XDG_CONFIG_HOME/nerdshade.yaml` or `$HOME/.config/nerdshade.yaml`.

Example `~/.config/nerdshade.yaml`:

```
latitude: 50.534
longitude: 8.6
tempnight: 4200
```

Command line parameters will take precedence over those from the configuration file.

You *can* add `loop: true` to the configuration file, although it could be confusing.


## Building

- Clone repository
- `make test` # optional
- `make`

## Running via systemd

An example of how you can run nerdshade through systemd:

```bash
systemd-run --user -u nerdshade nerdshade -nologtime -loop
```

The main benefit is to have the log output available in the system journal. To see it run:

```bash
journalctl --user -f -u nerdshade
```

To stop:

```bash
systemctl --user stop nerdshade
```

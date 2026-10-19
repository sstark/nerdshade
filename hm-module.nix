{ self }:
{ config, lib, pkgs, options, ... }:

let
  inherit (lib)
    mkEnableOption mkIf mkOption types literalExpression;
  cfg = config.services.nerdshade;
  format = pkgs.formats.yaml { };
  hyprsunsetSvc =
    cfg.hyprsunsetPackage != null
    && options ? services.hyprsunset
    && config.services.hyprsunset.enable;
  settings =
    lib.optionalAttrs cfg.debug { debug = true; }
    // lib.filterAttrs (n: v: v != null) {
      inherit (cfg)
        latitude longitude tempDay tempNight gammaDay gammaNight
        fixedWakeup fixedBedtime transitionDuration hyprctl;
    }
    // cfg.settings;
in
{
  options.services.nerdshade = {
    enable = mkEnableOption "nerdshade, a hyprsunset client that transitions color temperature during sunrise/sunset";

    package = mkOption {
      type = types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
      defaultText = literalExpression "nerdshade.packages.${pkgs.stdenv.hostPlatform.system}.default";
      description = "The nerdshade package to use.";
    };

    hyprsunsetPackage = mkOption {
      type = types.nullOr types.package;
      default = pkgs.hyprsunset;
      defaultText = literalExpression "pkgs.hyprsunset";
      description = ''
        The hyprsunset package to install. nerdshade only talks to a running
        hyprsunset instance via hyprctl, so hyprsunset needs to be available
        in the session (e. g. to start it with
        {command}`hyprctl keyword exec hyprsunset`). Set to null if you
        manage hyprsunset yourself.
      '';
    };

    latitude = mkOption {
      type = types.nullOr types.float;
      default = null;
      description = "Your location latitude. Used to calculate sunrise/sunset times.";
    };

    longitude = mkOption {
      type = types.nullOr types.float;
      default = null;
      description = "Your location longitude. Used to calculate sunrise/sunset times.";
    };

    tempDay = mkOption {
      type = types.nullOr types.ints.unsigned;
      default = null;
      description = "Day color temperature.";
    };

    tempNight = mkOption {
      type = types.nullOr types.ints.unsigned;
      default = null;
      description = "Night color temperature.";
    };

    gammaDay = mkOption {
      type = types.nullOr types.ints.unsigned;
      default = null;
      description = "Day gamma.";
    };

    gammaNight = mkOption {
      type = types.nullOr types.ints.unsigned;
      default = null;
      description = "Night gamma.";
    };

    fixedWakeup = mkOption {
      type = types.nullOr types.str;
      default = null;
      example = "6:00";
      description = ''
        Fixed wakeup time in 24-hour format. Overrides location-based
        calculation. Requires {option}`fixedBedtime` to also be set.
      '';
    };

    fixedBedtime = mkOption {
      type = types.nullOr types.str;
      default = null;
      example = "22:30";
      description = ''
        Fixed bedtime time in 24-hour format. Overrides location-based
        calculation. Requires {option}`fixedWakeup` to also be set.
      '';
    };

    transitionDuration = mkOption {
      type = types.nullOr types.str;
      default = null;
      example = "45m";
      description = ''
        Duration of the transition during sunrise/sunset, e. g. "45m" or
        "1h10m".
      '';
    };

    hyprctl = mkOption {
      type = types.nullOr types.str;
      default = null;
      description = "Path to the hyprctl program.";
    };

    debug = mkOption {
      type = types.bool;
      default = false;
      description = "Whether to print debug info.";
    };

    settings = mkOption {
      type = format.type;
      default = { };
      example = {
        latitude = 50.534;
        longitude = 8.6;
        tempNight = 4200;
      };
      description = ''
        Extra configuration written to
        {file}`$XDG_CONFIG_HOME/nerdshade.yaml`. Options set here take
        precedence over the dedicated options above.
      '';
    };

    systemd = {
      enable = mkOption {
        type = types.bool;
        default = true;
        description = "Whether to run nerdshade as a systemd user service.";
      };

      extraArgs = mkOption {
        type = types.listOf types.str;
        default = [ ];
        example = [ "-debug" ];
        description = "Additional command-line arguments passed to nerdshade.";
      };
    };
  };

  config = mkIf cfg.enable {
    assertions = [
      {
        assertion = (cfg.fixedWakeup == null) == (cfg.fixedBedtime == null);
        message = "services.nerdshade: both fixedWakeup and fixedBedtime need to be set, or neither.";
      }
    ];

    home.packages = [ cfg.package ]
      ++ lib.optional (cfg.hyprsunsetPackage != null) cfg.hyprsunsetPackage;

    xdg.configFile."nerdshade.yaml" = mkIf (settings != { }) {
      source = format.generate "nerdshade.yaml" settings;
    };

    systemd.user.services.nerdshade = mkIf cfg.systemd.enable {
      Unit = {
        Description = "nerdshade - hyprsunset brightness client";
        After = [ "graphical-session.target" ]
          ++ lib.optional hyprsunsetSvc "hyprsunset.service";
        Wants = lib.optional hyprsunsetSvc "hyprsunset.service";
        PartOf = [ "graphical-session.target" ];
        ConditionEnvironment = "WAYLAND_DISPLAY";
      };

      Service = {
        ExecStart = lib.escapeShellArgs ([ (lib.getExe cfg.package) "-loop" "-nologtime" ] ++ cfg.systemd.extraArgs);
        Restart = "on-failure";
        RestartSec = 5;
      };

      Install.WantedBy = [ "graphical-session.target" ];
    };
  };
}

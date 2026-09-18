{
  config,
  lib,
  pkgs,
  ...
}:

let
  cfg = config.services.h2static;

  port = lib.toInt (lib.last (lib.splitString ":" cfg.listenAddress));

  args = [
    "-addr"
    cfg.listenAddress
    "-dir"
    cfg.root
  ]
  ++ lib.optional cfg.logRequests "-log"
  ++ cfg.extraArgs;
in

{
  options.services.h2static = {
    enable = lib.mkEnableOption "the h2static web server";

    package = lib.mkPackageOption pkgs "h2static" { };

    listenAddress = lib.mkOption {
      type = lib.types.str;
      default = ":8080";
      example = "127.0.0.1:8080";
      description = "Address and port to listen on. The address can be omitted to bind all addresses.";
    };

    root = lib.mkOption {
      type = lib.types.path;
      example = "/srv/www";
      description = "Path of the root directory to serve.";
    };

    logRequests = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = "Whether to log requests.";
    };

    extraArgs = lib.mkOption {
      type = lib.types.listOf lib.types.str;
      default = [ ];
      example = [
        "-disable-index"
        "-show-dotfiles"
      ];
      description = "Extra command line arguments to pass to h2static.";
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.h2static = {
      description = "h2static web server";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      serviceConfig = {
        ExecStart = lib.escapeShellArgs ([ (lib.getExe cfg.package) ] ++ args);
        Restart = "on-failure";

        DynamicUser = true;
        AmbientCapabilities = lib.optional (port < 1024) "CAP_NET_BIND_SERVICE";
        CapabilityBoundingSet = lib.optional (port < 1024) "CAP_NET_BIND_SERVICE";

        BindReadOnlyPaths = [ cfg.root ];
        LockPersonality = true;
        MemoryDenyWriteExecute = true;
        PrivateDevices = true;
        ProtectClock = true;
        ProtectControlGroups = true;
        ProtectHome = true;
        ProtectHostname = true;
        ProtectKernelLogs = true;
        ProtectKernelModules = true;
        ProtectKernelTunables = true;
        ProtectProc = "invisible";
        RestrictAddressFamilies = [
          "AF_INET"
          "AF_INET6"
        ];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        SystemCallArchitectures = "native";
        SystemCallFilter = [
          "@system-service"
          "~@privileged @resources"
        ];
        UMask = "0077";
      };
    };
  };
}

# SPDX-FileCopyrightText: 2025 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  pkgs,
  lib,
  config,
  ...
}:
let
  cfg = config.services.cunicu;

  settingsFormat = pkgs.formats.yaml { };
in
{
  options.services.cunicu = {
    package = lib.mkPackageOption pkgs "cunicu" { };

    daemon = {
      enable = lib.mkEnableOption "cunicu mesh network daemon";

      settings = lib.mkOption {
        inherit (settingsFormat) type;

        description = ''
          cunicu configuration

          See: https://cunicu.li/docs/config
        '';
        default = { };
      };
    };

    signal = {
      enable = lib.mkEnableOption "cunicu signaling server";

      listen = lib.mkOption {
        description = "Listen address";
        type = lib.types.str;
        default = ":8080";
      };

      secure = lib.mkOption {
        description = "Listen with self-signed TLS certificate";
        type = lib.types.bool;
        default = false;
      };

      logLevel = lib.mkOption {
        description = "Log level";
        type = lib.types.str;
        default = "info";
      };
    };

    relay = {
      enable = lib.mkEnableOption "cunicu relay server";

      urls = lib.mkOption {
        description = "List of STUN & TURN servers";
        type = lib.types.listOf lib.types.str;
        default = [ ];
      };

      listen = lib.mkOption {
        description = "Listen address";
        type = lib.types.str;
        default = ":8080";
      };

      secure = lib.mkOption {
        description = "Listen with self-signed TLS certificate";
        type = lib.types.bool;
        default = false;
      };
    };
  };

  config = lib.mkIf (cfg.daemon.enable || cfg.relay.enable || cfg.signal.enable) {
    users = {
      users.cunicu = {
        home = "/var/lib/cunicu";
        isSystemUser = true;
        group = "cunicu";
      };

      groups.cunicu = { };
    };

    environment.etc."cunicu/cunicu.yaml" = {
      source = settingsFormat.generate "cunicu.yaml" cfg.daemon.settings;
    };

    environment.systemPackages = [ cfg.package ];

    systemd = {
      services = {
        cunicu = lib.mkIf cfg.daemon.enable {
          description = "cunīcu mesh network daemon";
          documentation = [ "https://cunicu.li/docs" ];

          wants = [ "network-online.target" ];
          after = [
            "network-online.target"
            "cunicu.socket"
          ];
          requires = [ "cunicu.socket" ];
          wantedBy = [ "multi-user.target" ];

          serviceConfig = {
            Type = "notify-reload";

            ExecStart = "${lib.getExe cfg.package} daemon --config /etc/cunicu/cunicu.yaml";

            DynamicUser = true;
            NotifyAccess = "main";
            WatchdogSec = 10;

            BindPaths = [
              "-/var/run/wireguard"
              "-/dev/net/tun"
            ];
            DeviceAllow = [
              "/dev/net/tun rw"
            ];

            RuntimeDirectory = [
              "cunicu"
              "wireguard"
            ];
            StateDirectory = [
              "cunicu"
            ];
            ConfigurationDirectory = [
              "cunicu"
            ];

            # Hardening
            AmbientCapabilities = [
              "CAP_NET_ADMIN"
              "CAP_NET_BIND_SERVICE"
              "CAP_SYS_MODULE"
            ];
            CapabilityBoundingSet = [
              "CAP_NET_ADMIN"
              "CAP_NET_BIND_SERVICE"
              "CAP_SYS_MODULE"
            ];

            LockPersonality = true;
            MemoryDenyWriteExecute = true;
            NoNewPrivileges = true;
            PrivateDevices = true;
            PrivateUsers = "self";
            PrivateMounts = true;
            PrivateTmp = true;
            ProcSubset = "pid";
            ProtectClock = true;
            ProtectControlGroups = true;
            ProtectHome = true;
            ProtectHostname = true;
            ProtectKernelLogs = true;
            ProtectKernelTunables = true;
            ProtectProc = "invisible";
            ProtectSystem = "strict";
            ReadWritePaths = [
              "-/etc/hosts"
            ];
            RestrictAddressFamilies = [
              "AF_UNIX"
              "AF_INET"
              "AF_INET6"
              "AF_NETLINK"
            ];
            RestrictNamespaces = true;
            RestrictRealtime = true;
            RestrictSUIDSGID = true;
            SystemCallFilter = "@system-service";
            SystemCallErrorNumber = "EPERM";
            SystemCallArchitectures = "native";
          };

          environment = {
            CUNICU_EXPERIMENTAL = "1";
            CUNICU_CONFIG_ALLOW_INSECURE = "1";
          };
        };

        cunicu-signal = lib.mkIf cfg.signal.enable {
          description = "cunicu signaling server";
          documentation = [ "https://cunicu.li/docs" ];

          wants = [ "network-online.target" ];
          after = [ "network-online.target" ];
          wantedBy = [ "multi-user.target" ];

          serviceConfig = {
            Type = "simple";
            User = "cunicu";
            Group = "cunicu";
            ExecStart =
              "${lib.getExe cfg.package} signal --log-level ${cfg.signal.logLevel} "
              + lib.cli.toGNUCommandLineShell { } { inherit (cfg.signal) secure listen; };

            # Hardening
            AmbientCapabilities = [
              "CAP_NET_ADMIN"
              "CAP_NET_BIND_SERVICE"
              "CAP_SYS_MODULE"
            ];
            CapabilityBoundingSet = [
              "CAP_NET_ADMIN"
              "CAP_NET_BIND_SERVICE"
              "CAP_SYS_MODULE"
            ];
          };
        };

        cunicu-relay = lib.mkIf cfg.relay.enable {
          description = "cunicu relay server";
          documentation = [ "https://cunicu.li/docs" ];

          wants = [ "network-online.target" ];
          after = [ "network-online.target" ];
          wantedBy = [ "multi-user.target" ];

          serviceConfig = {
            Type = "simple";
            User = "cunicu";
            Group = "cunicu";
            ExecStart =
              "${lib.getExe cfg.package} relay "
              + lib.cli.toGNUCommandLineShell { } { inherit (cfg.relay) secure listen; }
              + " "
              + builtins.concatStringsSep " " cfg.relay.urls;
          };
        };
      };

      sockets = {
        cunicu = lib.mkIf cfg.daemon.enable {
          description = "cunīcu mesh network daemon control socket";

          partOf = [ "cunicu.service" ];
          wantedBy = [ "sockets.target" ];

          socketConfig = {
            FileDescriptorName = "control";
            ListenStream = "%t/cunicu.sock";
            SocketUser = "root";
            SocketGroup = "cunicu";
            SocketMode = "0660";
            RemoveOnStop = true;
          };
        };
      };
    };
  };
}

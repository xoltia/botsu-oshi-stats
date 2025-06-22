{ pkgs, lib, config, inputs, ... }:

{
  packages = [ pkgs.templ ];
  languages.go.enable = true;
  services.postgres.enable = true;
  services.postgres.initialDatabases = [ { name = "botsu"; } ];
}

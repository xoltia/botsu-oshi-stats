{ pkgs, lib, config, inputs, ... }:

{
  packages = with pkgs; [
    templ
    tailwindcss_4
    tailwindcss-language-server
  ];
  languages.go.enable = true;
  services.postgres.enable = true;
  services.postgres.initialDatabases = [ { name = "botsu"; } ];

  scripts.build.exec = ''
    tailwindcss -i input.css -o static/output.css
    templ generate
    go build -o ./bin/server .
  '';
}

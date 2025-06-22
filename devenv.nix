{ pkgs, lib, config, inputs, ... }:

{
  packages = with pkgs; [
    templ
    tailwindcss_4
    tailwindcss-language-server
    entr
  ];
  languages.go.enable = true;
  services.postgres.enable = true;
  services.postgres.initialDatabases = [ { name = "botsu"; } ];

  scripts.build.exec = ''
    tailwindcss -i input.css -o static/styles.css
    templ generate
    go build -o ./bin/server .
  '';

  scripts.build-run.exec = ''
    build
    BOTSU_DB_URL="postgresql:///botsu" ./bin/server "$@"
  '';
  
  scripts.run-watch.exec = ''
    find . \
      \( -path '*/.*' -prune \) -o \
      \( -type f \
         \( -name '*.go'   ! -name '*_templ.go' \) \
         -o -name '*.templ' \
      \) -print \
    | entr -cr build-run "$@"
  '';
}

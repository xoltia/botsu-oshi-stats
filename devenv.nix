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

  scripts.generate.exec = ''
    tailwindcss -i server/static/_input.css -o server/static/tailwind.css
    templ generate
  '';

  scripts.build-server.exec = ''
    generate
    go build -o ./bin/server ./cmd/server
  '';
  
  scripts.build-run-server.exec = ''
    build-server
    BOTSU_DB_URL="postgresql:///botsu" ./bin/server "$@"
  '';
  
  scripts.watch-run-server.exec = ''
    find . \
      \( -path '*/.*' -prune \) -o \
      \( -type f \
         \( -name '*.go'   ! -name '*_templ.go' \) \
         -o -name '*.templ' \
      \) -print \
    | entr -cr build-run-server "$@"
  '';
}

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/xoltia/botsu-oshi-stats/server"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()
	s := server.Server{}
	err := http.ListenAndServe(*addr, s)
	if err != nil {
		log.Fatalln(err)
	}
}

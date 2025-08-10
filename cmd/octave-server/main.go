package main

import (
	"flag"
	"log"

	"github.com/fmcato/octave-mcp/internal/server"
)

var httpAddr = flag.String("http", "", "HTTP address to listen on (empty for stdio)")

func main() {
	flag.Parse()

	srv := server.New()
	srv.RegisterHandlers()

	if *httpAddr != "" {
		if err := srv.RunHTTP(*httpAddr); err != nil {
			log.Fatal(err)
		}
	}
	if err := srv.RunStdio(); err != nil {
		log.Fatal(err)
	}

}

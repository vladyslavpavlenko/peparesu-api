package main

import (
	"flag"
	"github.com/vladyslavpavlenko/peparesu/config"
	"log"
	"net/http"
)

var app config.AppConfig

func main() {
	err := setup(&app)
	if err != nil {
		log.Fatal()
	}

	addr := flag.String("addr", ":8080", "the api address")
	flag.Parse()

	log.Printf("Running on port %s", *addr)

	srv := &http.Server{
		Addr:    *addr,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

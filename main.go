package main

import (
	"log"
	"net/http"

	"pulley.com/shakesearch/api"
	"pulley.com/shakesearch/modules/searcher"
)

func main() {
	// Serves static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Loads file
	file := searcher.File{}
	err := file.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	// Listens
	api.Listen(file)
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"log"
	"net/http"
)

const port = ":8080"

func serve(dir string) error {
	log.Printf("Serving HTTP on http://0.0.0.0%s/ ...\n", port)
	return http.ListenAndServe(port, http.FileServer(http.Dir(dir)))
}

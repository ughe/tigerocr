package main

import (
	"log"
	"net/http"
)

const addr = "127.0.0.1:8080"

func serve(dir string) error {
	log.Printf("Serving HTTP on http://%s/.\n", addr)
	return http.ListenAndServe(addr, http.FileServer(http.Dir(dir)))
}

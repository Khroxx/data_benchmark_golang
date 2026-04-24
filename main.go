package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})
	addr := ":8080"
	log.Printf("Server läuft auf %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

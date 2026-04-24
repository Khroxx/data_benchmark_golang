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

	http.HandleFunc("/api/golang/benchmark", benchmarkHandler)

	addr := ":8080"
	log.Printf("Server läuft auf %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func benchmarkHandler(w http.ResponseWriter, r *http.Request) {
	typ := r.URL.Query().Get("type")
	sizeKb := r.URL.Query().Get("sizeKb")
	var data []byte
	switch typ {
	case "flat-json":
		data = generateFlatJSON(sizeKb)
	case "nested-json":
		data = generateNestedJSON(sizeKb)
	case "csv":
		data = generateCSV(sizeKb)
	case "blob":
		data = generateBlob(sizeKb)
	default:
		http.Error(w, "invalid type", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func generateFlatJSON(sizeKb string) []byte {
	return make([]byte, parseSize(sizeKb))
}
func generateNestedJSON(sizeKb string) []byte {
	return make([]byte, parseSize(sizeKb))
}
func generateCSV(sizeKb string) []byte {
	return make([]byte, parseSize(sizeKb))
}
func generateBlob(sizeKb string) []byte {
	return make([]byte, parseSize(sizeKb))
}
func parseSize(sizeKb string) int {
	var n int
	fmt.Sscanf(sizeKb, "%d", &n)
	return n * 1024
}

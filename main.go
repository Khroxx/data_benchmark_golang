package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const defaultRuns = 1

type benchmarkResponse struct {
	Type       string   `json:"type"`
	SizeKB     int      `json:"sizeKb"`
	Runs       int      `json:"runs"`
	Durations  []int64  `json:"durations"`
	AverageMS  float64  `json:"average_ms"`
	MedianMS   float64  `json:"median_ms"`
	DataBytes  int      `json:"data_bytes"`
	Generated  bool     `json:"generated"`
	ServerTime string   `json:"server_time"`
	Warnings   []string `json:"warnings,omitempty"`
}

func main() {
	loadEnvFile(".env")

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/api/golang/benchmark", benchmarkHandler)

	addr := ":" + firstNonEmpty(os.Getenv("PORT"), "8080")
	handler := withCORS(mux)
	log.Printf("Server laeuft auf %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "pong")
}

func benchmarkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	payloadType := strings.TrimSpace(query.Get("type"))
	if payloadType == "" {
		http.Error(w, "missing type query parameter", http.StatusBadRequest)
		return
	}

	sizeKB, err := parseSizeKB(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	runs, warnings := parseRuns(query.Get("runs"))
	durations := make([]int64, 0, runs)

	var payload []byte
	for i := 0; i < runs; i++ {
		start := time.Now()
		payload, err = generatePayload(payloadType, sizeKB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		durations = append(durations, time.Since(start).Milliseconds())
	}

	response := benchmarkResponse{
		Type:       payloadType,
		SizeKB:     sizeKB,
		Runs:       runs,
		Durations:  durations,
		AverageMS:  average(durations),
		MedianMS:   median(durations),
		DataBytes:  len(payload),
		Generated:  true,
		ServerTime: time.Now().UTC().Format(time.RFC3339),
		Warnings:   warnings,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func parseSizeKB(query url.Values) (int, error) {
	rawSize := strings.TrimSpace(firstNonEmpty(
		query.Get("sizeKb"),
		query.Get("size"),
	))
	if rawSize == "" {
		return 0, fmt.Errorf("missing sizeKb query parameter")
	}

	sizeKB, err := strconv.Atoi(rawSize)
	if err != nil || sizeKB < 1 {
		return 0, fmt.Errorf("invalid sizeKb query parameter")
	}

	return sizeKB, nil
}

func parseRuns(rawRuns string) (int, []string) {
	rawRuns = strings.TrimSpace(rawRuns)
	if rawRuns == "" {
		return defaultRuns, nil
	}

	runs, err := strconv.Atoi(rawRuns)
	if err != nil || runs < 1 {
		return defaultRuns, []string{"invalid runs value, defaulted to 1"}
	}

	return runs, nil
}

func generatePayload(payloadType string, sizeKB int) ([]byte, error) {
	targetBytes := sizeKB * 1024

	switch payloadType {
	case "flat-json":
		return buildFlatJSON(targetBytes), nil
	case "nested-json":
		return buildNestedJSON(targetBytes), nil
	case "csv":
		return buildCSV(targetBytes), nil
	case "blob":
		return buildBlob(targetBytes), nil
	default:
		return nil, fmt.Errorf("invalid type query parameter")
	}
}

func buildFlatJSON(targetBytes int) []byte {
	base := `{"id":1,"name":"benchmark-entry","status":"ok","category":"flat","active":true,"score":12345}`
	return padContent(base, targetBytes)
}

func buildNestedJSON(targetBytes int) []byte {
	base := `{"meta":{"name":"benchmark","version":1},"items":[{"id":1,"tags":["alpha","beta"],"payload":{"kind":"nested","enabled":true,"metrics":{"count":3,"value":42}}}]}`
	return padContent(base, targetBytes)
}

func buildCSV(targetBytes int) []byte {
	base := "id,name,status,value\n1,benchmark,ok,42\n2,runner,ok,84\n"
	return padContent(base, targetBytes)
}

func buildBlob(targetBytes int) []byte {
	base := "benchmark-payload-blob-"
	return padContent(base, targetBytes)
}

func padContent(base string, targetBytes int) []byte {
	if targetBytes <= 0 {
		return []byte{}
	}

	if len(base) >= targetBytes {
		return []byte(base[:targetBytes])
	}

	var builder strings.Builder
	builder.Grow(targetBytes)

	for builder.Len() < targetBytes {
		remaining := targetBytes - builder.Len()
		if remaining >= len(base) {
			builder.WriteString(base)
			continue
		}

		builder.WriteString(base[:remaining])
	}

	return []byte(builder.String())
}

func average(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}

	var total int64
	for _, value := range values {
		total += value
	}

	return float64(total) / float64(len(values))
}

func median(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return float64(sorted[middle])
	}

	return float64(sorted[middle-1]+sorted[middle]) / 2
}

func loadEnvFile(path string) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		_ = os.Setenv(key, value)
	}
}

func withCORS(next http.Handler) http.Handler {
	allowedOrigin := firstNonEmpty(os.Getenv("CORS_ALLOWED_ORIGIN"), "*")
	allowedMethods := firstNonEmpty(os.Getenv("CORS_ALLOWED_METHODS"), "GET, OPTIONS")
	allowedHeaders := firstNonEmpty(os.Getenv("CORS_ALLOWED_HEADERS"), "Content-Type, Authorization")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
}

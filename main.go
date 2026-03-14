package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/stats", handleStats)
	http.HandleFunc("/api/timeline", handleTimeline)
	http.Handle("/", http.FileServer(http.FS(staticFS)))

	fmt.Println("Claude Code Stats → http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleTimeline(w http.ResponseWriter, r *http.Request) {
	tl, err := ParseTimeline()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tl)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := ParseHistory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

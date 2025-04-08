package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h(w, r)
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	format := r.URL.Query().Get("format")

	if url == "" || (format != "mp3" && format != "mp4") {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("video_%d", timestamp)
	videoFile := filename + ".mp4"
	outputFile := filename + "." + format

	log.Println("Downloading video from:", url)

	cmd := exec.Command("yt-dlp", "-f", "best", "-o", videoFile, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Println("Error during yt-dlp:", err)
		http.Error(w, "Failed to download video", http.StatusInternalServerError)
		return
	}

	if format == "mp3" {
		log.Println("Converting to mp3...")
		cmd = exec.Command("ffmpeg", "-i", videoFile, "-b:a", "192K", "-vn", outputFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Println("Error during ffmpeg:", err)
			http.Error(w, "Failed to convert video", http.StatusInternalServerError)
			return
		}
		defer os.Remove(videoFile)
	} else {
		outputFile = videoFile
	}

	defer os.Remove(outputFile) 

	_ = filepath.Ext(outputFile)
	mimeType := "application/octet-stream"
	if format == "mp3" {
		mimeType = "audio/mpeg"
	} else if format == "mp4" {
		mimeType = "video/mp4"
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", outputFile))

	f, err := os.Open(outputFile)
	if err != nil {
		log.Println("Error opening file:", err)
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		log.Println("Error streaming file:", err)
		http.Error(w, "Failed to send file", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/download", withCORS(downloadHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

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
		http.Error(w, "Failed to download video", http.StatusInternalServerError)
		return
	}

	if format == "mp3" {
		log.Println("Converting to mp3...")
		cmd = exec.Command("ffmpeg", "-i", videoFile, "-b:a", "192K", "-vn", outputFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			http.Error(w, "Failed to convert video", http.StatusInternalServerError)
			return
		}
		defer os.Remove(videoFile)
	} else {
		outputFile = videoFile
	}

	defer os.Remove(outputFile)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", outputFile))
	w.Header().Set("Content-Type", "application/octet-stream")
	f, err := os.Open(outputFile)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}

func main() {
	http.HandleFunc("/download", downloadHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on port %s...", port)
	http.ListenAndServe(":"+port, nil)
}

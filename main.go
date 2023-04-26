package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	setEnvironment()
}

func main() {
	var directory = flag.String("dir", os.Getenv("DIRECTORY"), "serve directory")
	var address = flag.String("address", os.Getenv("ADDRESS"), "address to listen")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/api/playlist", func(w http.ResponseWriter, r *http.Request) {
		// Specify the directory containing the video files

		// Get a list of all the video files in the directory
		videoFiles, err := filepath.Glob(filepath.Join(*directory, "*.mp4"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create a list of playlist items
		playlist := make([]map[string]string, 0, len(videoFiles))
		for _, file := range videoFiles {
			playlist = append(playlist, map[string]string{
				"title":     filepath.Base(file),
				"video_url": "/videos/" + filepath.Base(file),
			})
		}

		// Encode the playlist data as JSON and write it to the response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(playlist); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Serve static files from videos directory
	http.Handle("/videos/", http.StripPrefix("/videos/", http.FileServer(http.Dir(*directory))))

	// Start the server
	http.ListenAndServe(*address, nil)
}

func setEnvironment() {
	// Open the file with the environment variables
	file, err := os.Open("env.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Scan the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Split the line by "="
		pair := strings.SplitN(scanner.Text(), "=", 2)
		if len(pair) != 2 {
			continue // Skip invalid lines
		}
		// Set the environment variable
		key := pair[0]
		value := pair[1]
		err := os.Setenv(key, value)
		if err != nil {
			log.Fatal(err)
		}
	}
}

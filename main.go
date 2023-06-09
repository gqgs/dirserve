package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed env.txt
//go:embed index.html
var content embed.FS

func init() {
	setEnvironment()
}

func main() {
	var directory = flag.String("dir", os.Getenv("DIRECTORY"), "serve directory")
	var address = flag.String("address", os.Getenv("ADDRESS"), "address to listen")
	var dev = flag.Bool("dev", false, "run in development mode")
	flag.Parse()

	if *dev {
		http.Handle("/", http.FileServer(http.FS(os.DirFS("."))))
	} else {
		http.Handle("/", http.FileServer(http.FS(content)))
	}

	http.HandleFunc("/api/playlist", func(w http.ResponseWriter, r *http.Request) {
		var playlist []map[string]string
		filepath.WalkDir(*directory, func(path string, d fs.DirEntry, err error) error {
			filename := d.Name()
			if isSupportedVideo(filename) {
				playlist = append(playlist, map[string]string{
					"title":     filename,
					"video_url": "/videos/" + filename,
				})
			}
			return nil
		})

		sort.Slice(playlist, func(i, j int) bool {
			return strings.ToLower(playlist[i]["title"]) < strings.ToLower(playlist[j]["title"])
		})

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
	file, err := content.Open("env.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Scan the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, found := strings.Cut(scanner.Text(), "=")
		if !found {
			continue
		}
		if os.Getenv(key) != "" {
			// already defined
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			log.Fatal(err)
		}
	}
}

func isSupportedVideo(name string) bool {
	return strings.HasSuffix(name, ".mp4") || strings.HasSuffix(name, ".webm")
}

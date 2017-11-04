package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"io/ioutil"
	"sort"
	"gopkg.in/russross/blackfriday.v2"
	"strings"
	"time"
)

func main() {
	// Get application folder
	app, err := os.Executable()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	appPath := filepath.Dir(app)

	staticPath := filepath.Join(appPath, "static")
	templatePath := filepath.Join(appPath, "templates")
	dataPath := filepath.Join(appPath, "data")

	// Configuration for serving static files
	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

	// Show activities
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		// Get activity files
		files, err := ioutil.ReadDir(dataPath)
		if err != nil {
			log.Fatal(err)
		}

		// Sort by file name (YYYYMMDD.md)
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() > files[j].Name()
		})

		type Activity struct {
			Date string
			Content template.HTML
		}

		// Create activity objects from files
		var activities []Activity
		for _, file := range files {
			// Parse activity date from filename
			time, err := time.Parse("20060102", file.Name()[:len(file.Name()) - 3])
			if err != nil {
				log.Fatal(err)
			}

			// Convert file content (markdown) to html
			path := filepath.Join(dataPath, file.Name())
			content, err := ioutil.ReadFile(path)
			if err != nil {
				log.Fatal(err)
			}
			content = []byte(strings.Replace(string(content), "\r\n", "\n", -1))
			contentHTML := blackfriday.Run(content)

			// Append activity to list
			activity := Activity{
				Date: time.Format("2006-01-02 (Mon)"),
				Content: template.HTML(string(contentHTML)),
			}
			activities = append(activities, activity)
		}

		// Render template
		t := template.Must(template.ParseFiles(
			filepath.Join(templatePath, "index.tpl")))
		err = t.ExecuteTemplate(w, "index.tpl", activities)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Start application
	port := os.Getenv("ACTIVITY_PORT")
	if port == "" {
		port = "80"
	}
	log.Fatal(http.ListenAndServe(":" + port, nil))
}

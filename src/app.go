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
	"fmt"
	"regexp"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"bytes"
)

var PASSWORD = []byte{
	230, 123, 62, 195, 18, 222, 4, 239, 56, 145, 231, 224, 103, 127, 247,
	105, 40, 87, 204, 238, 78, 85, 76, 54, 51, 176, 29, 81, 45, 92, 74, 232,
}
var SALT = "@ct1v1ty-"

type ActivityEditContext struct {
	New bool
	Date time.Time
	Title string
	Body string
	Error string
}

type GitHubFileContent struct {
	Content string `json:"content"`
	SHA string `json:"sha"`
}

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

	// Edit new activity
	http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		ctx := ActivityEditContext{
			New: true,
			Date: time.Now(),
			Title: "",
			Body: "",
			Error: "",
		}

		t := template.Must(template.ParseFiles(
			filepath.Join(templatePath, "edit.tpl")))
		err = t.ExecuteTemplate(w, "edit.tpl", ctx)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Create new activity
	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			fmt.Fprintf(w, http.StatusText(http.StatusMethodNotAllowed))
			return
		}

		r.ParseForm()

		date := r.Form["date"][0]
		title := r.Form["title"][0]
		body := r.Form["body"][0]
		password := r.Form["password"][0]

		retryEdit := func(date time.Time, title string, body string, reason string) {
			t := template.Must(template.ParseFiles(
				filepath.Join(templatePath, "edit.tpl")))
			err = t.ExecuteTemplate(w, "edit.tpl", ActivityEditContext{
				New: true,
				Date: date,
				Title: title,
				Body: body,
				Error: reason,
			})
			if err != nil {
				log.Fatal(err)
			}
		}

		// Check date
		re := regexp.MustCompile(`^\d{4}\d{2}\d{2}$`)
		if !re.MatchString(date) {
			retryEdit(time.Now(), title, body, "Invalid date format.")
			return
		}

		// Check password
		hash := sha256.Sum256([]byte(SALT + password))
		for i := range hash {
			if PASSWORD[i] != hash[i] {
				date, err := time.Parse("20060102", string(date))
				if err != nil {
					date = time.Now()
				}
				retryEdit(date, title, body, "Invalid password.")
				return
			}
		}

		path := filepath.Join(dataPath, date + ".md")
		body = "# " + title + "\n\n" + body

		// Add or update activity
		contentOnGitHub := getActivityFromGitHub(date)
		if contentOnGitHub == nil {
			// Add new activity
			commitActivityToGitHub(date, body)
		} else {
			// Update existing activity
			content, err := base64.StdEncoding.DecodeString(contentOnGitHub.Content)
			if err != nil {
				log.Fatal(err)
			}
			body = string(content) + "\n\n" + body
			updateActivityOnGitHub(date, body, contentOnGitHub.SHA)
		}

		// Save content to local file
		err = ioutil.WriteFile(path, []byte(body), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		http.Redirect(w, r, "/", http.StatusFound)
	})

	// Start application
	port := os.Getenv("ACTIVITY_PORT")
	if port == "" {
		port = "80"
	}
	log.Fatal(http.ListenAndServe(":" + port, nil))
}

func getActivityFromGitHub(date string) *GitHubFileContent {
	// Set API endpoint
	url := fmt.Sprintf(
		"https://tsujio:%s@api.github.com/repos/tsujio/www-activity/contents/src/data/%s.md",
		os.Getenv("GITHUB_TOKEN"),
		date,
	)

	// Call API
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Parse json
	var content GitHubFileContent
	err = json.Unmarshal(b, &content)
	if err != nil {
		log.Fatal(err)
	}

	return &content
}

func updateActivityOnGitHub(date string, body string, sha string) {
	// Set API endpoint
	url := fmt.Sprintf(
		"https://tsujio:%s@api.github.com/repos/tsujio/www-activity/contents/src/data/%s.md",
		os.Getenv("GITHUB_TOKEN"),
		date,
	)

	// Create request body
	json, err := json.Marshal(&struct {
		Message string `json:"message"`
		Content string `json:"content"`
		SHA string `json:"sha"`
	} {
		"Update activity",
		base64.StdEncoding.EncodeToString([]byte(body)),
		sha,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create request
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(json))
	if err != nil {
		log.Fatal(err)
	}

	// Call API
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(string(b))
	}
}

func commitActivityToGitHub(date string, body string) {
	// Set API endpoint
	url := fmt.Sprintf(
		"https://tsujio:%s@api.github.com/repos/tsujio/www-activity/contents/src/data/%s.md",
		os.Getenv("GITHUB_TOKEN"),
		date,
	)

	// Create request body
	json, err := json.Marshal(&struct {
		Message string `json:"message"`
		Content string `json:"content"`
	} {
		"Add activity",
		base64.StdEncoding.EncodeToString([]byte(body)),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create request
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(json))
	if err != nil {
		log.Fatal(err)
	}

	// Call API
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(string(b))
	}
}

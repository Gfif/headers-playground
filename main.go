package main

import (
	"encoding/json"
	"errors"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	text "text/template"
)

var logger = log.New(os.Stdout, "", log.Lshortfile)
var auth = flag.String("auth", "", "<user:password>")
var port = flag.String("port", "80", "<port>")

var user, pass string
var isAuth bool

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type Page struct {
	ID      string `json:"id"`
	Headers string `json:"headers"`
	Body    string `json:"body"`
}

func NewPage(id, headers, body string) *Page {
	return &Page{id, headers, body}
}

func LoadPage(id string) (*Page, error) {
	filename := "pages/" + id + ".json"
	body, _ := ioutil.ReadFile(filename)
	page := new(Page)
	err := json.Unmarshal(body, page)
	return page, err
}

func (p *Page) save() error {
	filename := "pages/" + p.ID + ".json"
	jsn, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, jsn, 0600)
}

func handler(w http.ResponseWriter, r *http.Request) {

	if isAuth {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"\"")
		if username, password, ok := r.BasicAuth(); !ok {
			http.Error(w, "Not Authorized", http.StatusUnauthorized)
			return
		} else {
			if username != user || password != pass {
				http.Error(w, "Not Authorized", http.StatusUnauthorized)
				return
			}
		}
	}

	if r.Method == "GET" {
		id := r.URL.Path[1:]
		if id == "" {
			t, _ := template.ParseFiles("index.html")
			t.Execute(w, nil)
		} else {
			p, err := LoadPage(id)
			if err != nil {
				logger.Printf("Error: %s", err.Error())
				http.NotFound(w, r)
			}

			for _, line := range strings.Split(p.Headers, "\n") {
				if strings.Contains(line, ":") {
					parts := strings.Split(line, ":")
					w.Header().Add(parts[0], strings.Join(parts[1:], ":"))
				}
			}
			t, _ := text.New("page").Parse("{{.Body}}")
			t.Execute(w, p)

		}
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			logger.Printf("Error: %s", err.Error())
			http.NotFound(w, r)
		}

		p := NewPage(r.Form["id"][0], r.Form["headers"][0], r.Form["body"][0])
		if err := p.save(); err != nil {
			logger.Printf("Error: %s", err.Error())
			http.NotFound(w, r)
		}
		http.Redirect(w, r, "/"+p.ID, http.StatusMovedPermanently)
	} else {
		http.Error(w, "Nooooooo!!!", http.StatusMethodNotAllowed)
	}
}

func main() {
	flag.Parse()

	if *auth != "" {
		isAuth = true

		parts := strings.Split(*auth, ":")
		if len(parts) != 2 {
			logger.Fatal(errors.New("Wrong auth var format"))
		}

		user, pass = parts[0], parts[1]
	}

	http.HandleFunc("/", handler)
	logger.Printf("Starting server on port %s", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		logger.Fatal(err)
	}
}

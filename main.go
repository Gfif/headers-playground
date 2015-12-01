package main

import (
	"encoding/json"
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

func LoadPage(id string) *Page {
	filename := "pages/" + id + ".json"
	body, _ := ioutil.ReadFile(filename)
	page := new(Page)
	_ = json.Unmarshal(body, page)
	return page
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
	if r.Method == "GET" {
		id := r.URL.Path[1:]
		if id == "" {
			t, _ := template.ParseFiles("index.html")
			t.Execute(w, nil)
		} else {
			p := LoadPage(id)
			for _, line := range strings.Split(p.Headers, "\n") {
				if strings.Contains(line, ":") {
					parts := strings.Split(line, ":")
					w.Header().Add(parts[0], parts[1])
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
		http.NotFound(w, r)
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8000", nil)
}

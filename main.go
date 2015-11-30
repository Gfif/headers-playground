package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
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

func NewPage(headers, body string) *Page {
	id := RandStringBytes(5)
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
			fmt.Println(p)
		}
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			logger.Printf("Error: %s", err.Error())
			http.NotFound(w, r)
		}

		p := NewPage(r.Form["headers"][0], r.Form["body"][0])
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
	http.ListenAndServe(":8080", nil)
}

// Copyright (c) 2014, Daniel Gallagher
// Use of this source code is covered by the MIT License, the full
// text of which can be found in the LICENSE file.

// Package main implements the wiki.  It creates the HTML templates
// for viewing, editing, and searching for wiki pages, and uses the
// wikilang package to convert from wiki source to HTML.  See the
// README and the initial wiki pages for information on how to use and
// setup the wiki.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/danielgallagher0/docwiki/wikilang"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

// Page is a container for wiki pages.  The fields are exported so
// that they can be used to fill in templates.  Note that *Page also
// defines a PrettyTitle() function that is used by the templates.
type Page struct {
	Title string
	Body  []byte
}

const viewPath = "/view/"
const editPath = "/edit/"
const savePath = "/save/"
const searchPath = "/search/"
const docPath = "/doc/"

const dataDir = "data/"
const tmplDir = "tmpl/"

const titleRegexp = "[A-Za-z0-9]+"

var templates = template.Must(template.ParseFiles(tmplDir+"edit.html",
	tmplDir+"view.html",
	tmplDir+"search.html"))
var titleValidator = regexp.MustCompile("^" + titleRegexp + "$")

var proxyRootPath string

func SetProxyRoot(p string) {
	proxyRootPath = p
}

func proxyRoot() string {
	return proxyRootPath
}

func (p *Page) PrettyTitle() string {
	return wikilang.WikiCase(p.Title)
}

func (p *Page) ProxyRoot() string {
	return proxyRoot()
}

func renderTemplate(w http.ResponseWriter, file string, p *Page) {
	var buf bytes.Buffer

	err := templates.ExecuteTemplate(&buf, file+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body := buf.Bytes()

	if file == "view" || file == "search" {
		body = []byte(wikilang.WikiToHtml(string(body)))
	}

	w.Write(body)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, proxyRoot()+editPath+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, proxyRoot()+viewPath+title, http.StatusFound)
}

func searchHandler(w http.ResponseWriter, r *http.Request, title string) {
	matches := make(chan string)
	count := 0

	filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		count++
		go func(file, target string) {
			match := ""

			f, err := os.Open(file)
			if err != nil {
				matches <- ""
			}

			re := regexp.MustCompile(fmt.Sprintf("\\[%s\\]", target))
			s := bufio.NewScanner(f)
			for s.Scan() {
				str := s.Text()

				if len(re.FindString(str)) > 0 {
					match = file[len(dataDir):]
					match = match[0 : len(match)-4]
					break
				}
			}

			matches <- match
		}(path, title)

		return nil
	})

	var body []byte
	for i := 0; i < count; i++ {
		match := <-matches
		if len(match) > 0 {
			body = append(body, []byte(fmt.Sprintf("- [%s]\n", match))...)
		}
	}
	p := &Page{Title: title, Body: body}

	renderTemplate(w, "search", p)
}

func loadPage(title string) (*Page, error) {
	filename := dataDir + url.QueryEscape(title) + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func (p *Page) save() error {
	filename := dataDir + url.QueryEscape(p.Title) + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func makeHandler(handler func(http.ResponseWriter, *http.Request, string), path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path[len(path):]
		if !titleValidator.MatchString(title) {
			http.NotFound(w, r)
			return
		}
		handler(w, r, title)
	}
}

func redirectToFrontPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, proxyRoot()+viewPath+"FrontPage", http.StatusFound)
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func ListenAndServe(port int) {
	http.HandleFunc("/", redirectToFrontPage)
	http.HandleFunc(viewPath, makeHandler(viewHandler, viewPath))
	http.HandleFunc(editPath, makeHandler(editHandler, editPath))
	http.HandleFunc(savePath, makeHandler(saveHandler, savePath))
	http.HandleFunc(searchPath, makeHandler(searchHandler, searchPath))
	http.HandleFunc(docPath, fileHandler)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}

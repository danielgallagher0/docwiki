package wikilang

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

const projectIndex = "projectIndex.xml"

type ProjectIndex struct {
	urls     map[string]string
	done     bool
	notifier chan bool
}

var projectDocs map[string]*ProjectIndex

func init() {
	projectDocs = make(map[string]*ProjectIndex)

	type Project struct {
		Name       string `xml:"name,attr"`
		SearchData string `xml:"searchdata"`
	}
	type Result struct {
		Project []Project `xml:"project"`
	}

	data, err := ioutil.ReadFile(projectIndex)
	if err != nil {
		panic(fmt.Sprintf("Could not read %s", projectIndex))
	}

	var result Result
	if err = xml.Unmarshal(data, &result); err != nil {
		panic(fmt.Sprintf("Could not read %s: %s", projectIndex, err))
	}

	for _, project := range result.Project {
		indexer := &ProjectIndex{map[string]string{}, false, make(chan bool)}
		projectDocs[project.Name] = indexer
		go indexer.Index(project.SearchData)
	}
}

func DocLink(project, entity string) string {
	url := "index.html"
	indexer, ok := projectDocs[project]
	if ok {
		if !indexer.done {
			indexer.done = <-indexer.notifier
		}
		url, _ = indexer.urls[entity]
	}

	return "/doc/" + project + "/html/" + url
}

func (indexer *ProjectIndex) Index(searchData string) {
	type Field struct {
		Name  string `xml:"name,attr"`
		Value string `xml:",innerxml"`
	}
	type Doc struct {
		Fields []Field `xml:"field"`
	}
	type Result struct {
		Docs []Doc `xml:"doc"`
	}

	data, err := ioutil.ReadFile(searchData)
	if err != nil {
		panic(fmt.Sprintf("Could not read %s", searchData))
	}

	var result Result
	if err = xml.Unmarshal(data, &result); err != nil {
		panic(fmt.Sprintf("Could not read %s: %s", searchData, err))
	}

	for _, doc := range result.Docs {
		name := ""
		url := ""
		for _, field := range doc.Fields {
			if field.Name == "name" {
				name = field.Value
			} else if field.Name == "url" {
				url = field.Value
			}
		}

		if len(name) > 0 && len(url) > 0 {
			indexer.urls[name] = url
		}
	}

	indexer.notifier <- true
}

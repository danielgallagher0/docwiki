// Copyright (c) 2014, Daniel Gallagher
// Use of this source code is covered by the MIT License, the full
// text of which can be found in the LICENSE file.

package wikilang

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

const projectIndexFile = "projectIndex.xml"

type projectIndex struct {
	urls     map[string]string
	done     bool
	notifier chan bool
}

var projectDocs map[string]*projectIndex

func init() {
	projectDocs = make(map[string]*projectIndex)

	type Project struct {
		Name       string `xml:"name,attr"`
		SearchData string `xml:"searchdata"`
	}
	type Result struct {
		Project []Project `xml:"project"`
	}

	data, err := ioutil.ReadFile(projectIndexFile)
	if err != nil {
		panic(fmt.Sprintf("Could not read %s", projectIndexFile))
	}

	var result Result
	if err = xml.Unmarshal(data, &result); err != nil {
		panic(fmt.Sprintf("Could not read %s: %s", projectIndexFile, err))
	}

	for _, project := range result.Project {
		indexer := &projectIndex{map[string]string{}, false, make(chan bool)}
		projectDocs[project.Name] = indexer
		go indexer.index(project.SearchData)
	}
}

// DocLink searches the indexed project to find the URL for a
// particular entity in the project's Doxygen.  The entity can be
// anything that Doxygen provides a link to, such as classes, methods,
// functions, types, etc.
//
// If the project or entity does not exist, the URL will be to the
// project's Doxygen index.  It may or may not exist.
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

func (indexer *projectIndex) index(searchData string) {
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

// Copyright (c) 2014, Daniel Gallagher
// Use of this source code is covered by the MIT License, the full
// text of which can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const confFile = "docwiki.conf"

func main() {
	type Config struct {
		Port      int
		ProxyRoot string
	}

	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		panic(fmt.Sprintf("Could not read %s: %s", confFile, err))
	}

	var conf Config
	if err = json.Unmarshal(data, &conf); err != nil {
		panic(fmt.Sprintf("Could not read %s: %s", confFile, err))
	}

	SetProxyRoot(conf.ProxyRoot)
	ListenAndServe(conf.Port)
}

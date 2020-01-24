package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
)

func main() {
	dir := os.Getenv("EXAMPLES_DIRECTORY")
	if dir == "" {
		panic("Missing `EXAMPLES_DIRECTORY` environment variable")
	}

	subdirs, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	data := make([]Data, 0)
	for _, subdir := range subdirs {

		if !subdir.IsDir() {
			continue
		}

		files, err := ioutil.ReadDir(path.Join(dir, subdir.Name()))
		if err != nil {
			panic(err)
		}

		for _, f := range files {
			if !strings.HasSuffix(f.Name(), ".graphql") {
				continue
			}
			withoutExt := strings.Replace(f.Name(), ".graphql", "", -1)
			withoutDash := strings.Replace(withoutExt, "-", " ", -1)
			data = append(data, Data{Description: withoutDash, FileName: f.Name(), Category: subdir.Name()})
		}
	}
	t, err := template.ParseFiles("./md.tpl")
	if err != nil {
		panic(err)
	}

	dest, err := os.Create(fmt.Sprintf("%s/README.md", dir))
	if err != nil {
		panic(err)
	}
	defer func() {
		err := dest.Close()
		if err != nil {
			panic(err)
		}
	}()
	err = t.Execute(dest, data)
	if err != nil {
		panic(err)
	}
}

type Data struct {
	FileName    string
	Description string
	Category    string
}

package web

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func ListenAndServe() {
	fileServer := http.FileServer(http.Dir("."))
	http.HandleFunc("/", renderIndex)
	http.Handle("/files/", http.StripPrefix("/files/", fileServer))
	if err := http.ListenAndServe("0.0.0.0:6004", nil); err != nil {
		fmt.Println("HERE")
		log.Fatal(err)
	}
}

func renderIndex(w http.ResponseWriter, r *http.Request) {
	files := []string{}

	dirContents, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range dirContents {
		if strings.HasSuffix(file.Name(), ".txt") {
			files = append(files, file.Name())
		}
	}
	vars := struct{ Files []string }{Files: files}

	parsedTemplate, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}
	if err = parsedTemplate.Execute(w, vars); err != nil {
		log.Fatal(err)
	}
}

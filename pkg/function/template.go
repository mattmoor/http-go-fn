package function

import "html/template"

const packageMain = `
package main

import (
       "fmt"
	"log"
	"net/http"
	"os"

        p "{{.Package}}"
)

func main() {
	http.HandleFunc("/", p.{{.Function}})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
`

var mainTemplate = template.Must(template.New("http-go-function-main").Parse(packageMain))

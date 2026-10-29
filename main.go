package main

import (
	"fmt"
	"log"
	"net/http"
)

// version 由构建时注入，如 -ldflags "-X main.version=v1.0.0"
var version = "dev"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello from go-demo, version: %s\n", version)
	})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

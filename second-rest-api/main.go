package main

import (
	"net/http"
	"fmt"
)

func main() {
	http.HandleFunc("/", handler)
	err := http.ListenAndServe("localhost:7070", nil)
	if err != nil {
		fmt.Println(err)
	}
}


func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world")
}

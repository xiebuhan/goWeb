package main

import (
	"fmt"
	"net/http"
)

func handlefunc(w http.ResponseWriter, r *http.Request)  {
	fmt.Fprint(w, "<h1>Hello, 这里是 goblog</h1>")
}

func main() {
	http.HandleFunc("/", handlefunc)
	http.ListenAndServe(":3000", nil)
}

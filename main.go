package main

import (
	"fmt"
	"net/http"
	"os"

	table2svg "github.com/k-p5w/go-table2svg/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}
	fmt.Println("hey")
	http.HandleFunc("/", table2svg.Handler)
	http.ListenAndServe(":"+port, nil)
}

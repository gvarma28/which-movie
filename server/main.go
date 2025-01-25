package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"github.com/gvarma28/which-movie/server/extractor"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Println(ctx)
	fmt.Printf("got / request\n")
	io.WriteString(w, "This is my website!\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /hello request\n")
	url := r.URL.Query().Get("url")
	test, err := extractor.GetComments(url)
	if err != nil {
		fmt.Printf("invalid response from GetComments\n")
	}
	io.WriteString(w, fmt.Sprintf("Hello, HTTP!\n %s %s", url, test))
}

func main() {
	http.HandleFunc("/", getRoot)
	http.HandleFunc("/hello", getHello)

	fmt.Printf("server has started\n")
	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

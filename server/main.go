package main

import (
	"errors"
	"fmt"
	"github.com/gvarma28/which-movie/server/extractor"
	"io"
	"net/http"
	"encoding/json"
	"os"
)

type Response struct {
	Message []string `json:"message"`
	Success bool   `json:"success"`
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!\n")
}

func doMagic(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	test, err := extractor.GetComments(url)
	if err != nil {
		fmt.Printf("invalid response from GetComments\n")
	}

	response := Response{
		Message: test,
		Success: true,
	}
	// Encode the struct to JSON and send it in the response body
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If encoding fails, return an error response
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// io.WriteString(w, fmt.Sprintf("%s \n\n %s", url, test))
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173") // Allow your SvelteKit app
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight (OPTIONS) requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	const PORT = ":8080"
	mux := http.NewServeMux()

	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/magic", doMagic)

	fmt.Printf("server has started at port %s\n", PORT)
	err := http.ListenAndServe(PORT, enableCors(mux))
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

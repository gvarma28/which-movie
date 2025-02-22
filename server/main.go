package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gvarma28/which-movie/server/extractor"
	"github.com/gvarma28/which-movie/server/processor"
	"github.com/joho/godotenv"
)

type Response struct {
	Results []processor.MovieResult `json:"results"`
	Success bool                    `json:"success"`
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!\n")
}

func doMagic(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	extractedData, err := extractor.ExtractData(url)
	if err != nil {
		fmt.Printf("invalid response from GetComments\n")
		response := Response{
			Success: false,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			// If encoding fails, return an error response
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	var magicResult processor.MagicResult
	comments := extractedData.Comments
	if extractedData.Title != nil {
		comments = append(comments, *extractedData.Title)
	}
	if extractedData.Subtitles != nil {
		comments = append(comments, *extractedData.Subtitles)
	}
	magicResult, err = processor.ProcessExtractedComments(comments)
	if err != nil {
		fmt.Printf("invalid response from ProcessExtractedComments - processAll\n")
	}

	response := Response{
		Results: magicResult.Results,
		Success: true,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "https://which-movie.com")
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
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error loading .env file\n")
	}
	const PORT = "0.0.0.0:8080"
	mux := http.NewServeMux()

	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/magic", doMagic)

	fmt.Printf("server has started at port %s\n", PORT)
	err = http.ListenAndServe(PORT, enableCors(mux))
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

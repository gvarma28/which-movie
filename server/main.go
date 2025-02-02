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
	Message []string `json:"message,omitempty"`
	Result  string   `json:"result"`
	Success bool     `json:"success"`
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!\n")
}

const (
	processingComments  = false
	processingSubtitles = false
	processAll          = true
)

func doMagic(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	extractedData, err := extractor.ExtractData(url)
	if err != nil {
		fmt.Printf("invalid response from GetComments\n")
		response := Response{
			Result:  "Unknown",
			Success: false,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			// If encoding fails, return an error response
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	var movieName string
	// process subtitles
	if processAll {
		comments := extractedData.Comments
		comments = append(comments, *extractedData.Title, *extractedData.Subtitles)
		movieNameAddr, err := processor.ProcessExtractedComments(comments)
		if err != nil {
			fmt.Printf("invalid response from ProcessExtractedComments - processAll\n")
		}
		movieName = *movieNameAddr
	} else if processingComments {
		movieNameAddr, err := processor.ProcessExtractedComments(extractedData.Comments)
		if err != nil {
			fmt.Printf("invalid response from ProcessExtractedComments - processingComments\n")
		}
		movieName = *movieNameAddr
	} else if processingSubtitles {
		movieNameAddr, err := processor.ProcessExtractedSubtitles(*extractedData.Subtitles)
		if err != nil {
			fmt.Printf("invalid response from ProcessExtractedSubtitles - processAll\n")
		}
		movieName = *movieNameAddr
	}

	response := Response{
		Result:  movieName,
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
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error loading .env file\n")
	}
	const PORT = ":8080"
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

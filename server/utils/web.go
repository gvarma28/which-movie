package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func GetRequest(url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error while creating the GET request: %v\n", err)
		return nil, errors.New("error while creating the GET request")
	}
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	client := &http.Client{}
	// Send the request
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making the GET request: %v\n", err)
		return nil, errors.New("error making the GET request")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading the response body: %v\n", err)
		return nil, errors.New("error reading the response body")
	}
	return body, nil
}

func PostRequest(url string, headers map[string]string, data []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Error while creating the POST request: %v\n", err)
		return nil, errors.New("error while creating the GET request")
	}
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	client := &http.Client{}
	// Send the request
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making the POST request: %v\n", err)
		return nil, errors.New("error making the POST request")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading the response body: %v\n", err)
		return nil, errors.New("error reading the response body")
	}

	return body, nil
}
package processor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type ProcessType string

const (
	Comments  ProcessType = "Comments"
	Subtitles ProcessType = "Subtitles"
)

func ProcessExtractedComments(extractedData []string) (*string, error) {

	body, err := makeRequest(extractedData, Comments)
	if err != nil {
		fmt.Printf("error while processing extracted comments %v\n", err)
	}
	var response APIResponse
	err = json.Unmarshal([]byte(*body), &response)
	if err != nil {
		fmt.Printf("error parsing JSON: %s\n", err)
	}

	return &response.Choices[0].Message.Content, nil
}

func ProcessExtractedSubtitles(extractedData string) (*string, error) {

	body, err := makeRequest(extractedData, Subtitles)
	if err != nil {
		fmt.Printf("error while processing extracted subtitles %v\n", err)
	}
	var response APIResponse
	err = json.Unmarshal([]byte(*body), &response)
	if err != nil {
		fmt.Printf("error parsing JSON: %s\n", err)
	}

	return &response.Choices[0].Message.Content, nil
}


func getRequestBody(extractedData any, processType ProcessType) ([]byte, error) {

	var combinedData string
	switch v := extractedData.(type) {
	case []string:
		combinedData = strings.Join(v, "\n")
	case string:
		combinedData = v
	}

	var messages []Messages
	switch processType {
	case Comments:
		messages = []Messages{
			{
				Role:    "system",
				Content: "You are a movie geek and know everything about movies and shows.",
			},
			{
				Role:    "user",
				Content: "I will give you a list of comments, analyse them and output the possible movie/tv show that the comments talk about. Give me just name of the possible movies. Can you do that?",
			},
			{
				Role:    "user",
				Content: combinedData,
			},
		}
	case Subtitles:
		{
			messages = []Messages{
				{
					Role:    "system",
					Content: "You are a movie geek and know everything about movies and shows.",
				},
				{
					Role:    "user",
					Content: "I will give you the subtitles of a youtube short. Please analyse it and output the possible movie/tv show that the subtitles talk about. Give me just name of the possible movies. Can you do that?",
				},
				{
					Role:    "user",
					Content: combinedData,
				},
			}
		}

	}

	requestBody := RequestBody{
		Model:    "gpt-4o-mini",
		Messages: messages,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, errors.New("error parsing requestBody - getRequestBody")
	}
	return jsonData, nil
}

func makeRequest(extractedData any, processType ProcessType) (*string, error) {

	url := "https://api.openai.com/v1/chat/completions"
	method := "POST"

	jsonData, err := getRequestBody(extractedData, processType)
	if err != nil {
		fmt.Printf("error while preparing the request body, err: %s\n", err)
		return nil, errors.New("error preparing the request body")
	}
	reader := strings.NewReader(string(jsonData))

	open_ai_token := os.Getenv("OPEN_AI_TOKEN")
	client := &http.Client{}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		fmt.Printf("error while creating the POST request, err: %s\n", err)
		return nil, errors.New("error while creating the POST request - GetMovieName")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", open_ai_token))
	// req.Header.Add("Cookie", "__cf_bm=HxznO2ypQg67tzvZwFb91rtwVzizE.ezEp0eQQANZr8-1738251971-1.0.1.1-5DjFtcss0myfe2JeN8JfhIRTFza2Blk049ysSKQ_nHEMGzCwtpUQGbpR4OLKS6PWyTzMc2uHTIP46nY2Q.KlAg; _cfuvid=REI_s1LJnYSjfX7UOz08L89c1tayu_DwpY.BlqoYaqE-1738251971613-0.0.1.1-604800000")

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("error making the POST request, err: %s\n", err)
		return nil, errors.New("error making the POST request - GetMovieName")
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("error reading the response body, err: %s\n", err)
		return nil, errors.New("error reading the response body - GetMovieName")
	}

	bodyStr := string(body)

	return &bodyStr, nil
}

// // UnmarshalResponse attempts to unmarshal the response body into the appropriate type
// func UnmarshalResponse(body []byte) (interface{}, error) {
//     // First, try to unmarshal into a map to check for error field
//     var raw map[string]interface{}
//     if err := json.Unmarshal(body, &raw); err != nil {
//         return nil, fmt.Errorf("failed to parse JSON: %v", err)
//     }
//     // Check if the response contains an error field
//     if _, hasError := raw["error"]; hasError {
//         var errorResp APIErrorResponse
//         if err := json.Unmarshal(body, &errorResp); err != nil {
//             return nil, fmt.Errorf("failed to parse error response: %v", err)
//         }
//         return errorResp, nil
//     }
//     // If no error field, treat as successful response
//     var successResp APIResponse
//     if err := json.Unmarshal(body, &successResp); err != nil {
//         return nil, fmt.Errorf("failed to parse success response: %v", err)
//     }
//     return successResp, nil
// }

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

func ProcessExtractedComments(extractedData []string) (MagicResult, error) {

	body, err := makeRequest(extractedData)
	if err != nil {
		fmt.Printf("error while processing extracted comments %v\n", err)
	}
	var response APIResponse
	err = json.Unmarshal([]byte(*body), &response)
	if err != nil {
		fmt.Printf("error parsing JSON: %s\n", err)
	}

	content := response.Choices[0].Message.Content
	var magicResult MagicResult
	err = json.Unmarshal([]byte(content), &magicResult)
	if err != nil {
		fmt.Printf("error parsing JSON: %s\n", err)
	}

	return magicResult, nil
}

func getRequestBody(extractedData any) ([]byte, error) {

	var combinedData string
	switch v := extractedData.(type) {
	case []string:
		combinedData = strings.Join(v, "\n")
	case string:
		combinedData = v
	}

	messages := []Messages{
		{
			Role:    "system",
			Content: "You are a film expert. Return only JSON data about movies/shows mentioned in user comments.",
		},
		{
			Role:    "user",
			Content: "Format required:\n{\n  \"results\": [\n    {\n      \"movie_name\": \"title\",\n      \"year\": YYYY,\n      \"short_description\": \"50-word max plot summary\"\n    }\n  ]\n}\nRules: List most confident matches first. Include year only if multiple versions exist. Keep descriptions brief.",
		},
		{
			Role:    "user",
			Content: combinedData,
		},
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

func makeRequest(extractedData any) (*string, error) {

	url := "https://api.openai.com/v1/chat/completions"
	method := "POST"

	jsonData, err := getRequestBody(extractedData)
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

// messages = []Messages{
// 	{
// 		Role:    "system",
// 		Content: "You are a highly knowledgeable film and television expert. Your task is to analyze user comments and identify which movies or TV shows they are discussing. You should consider plot points, character names, iconic scenes, and contextual clues in the comments to make accurate identifications.",
// 	},
// 	{
// 		Role: "user",
// 		Content: `Analyze the following user comments and identify the movies or TV shows being discussed.

// Requirements:
// - Return ONLY movie/show titles in a comma-separated list
// - Order titles from most to least likely based on comment relevance
// - If multiple titles fit equally well, list all of them
// - If a comment could refer to both a movie and its remake, include both versions
// - Include the year for movies with the same title (e.g., "Dune (1984), Dune (2021)")

// Example output: "The Dark Knight (2008), Batman Begins (2005), The Batman (2022)"`,
// 	},
// 	{
// 		Role:    "user",
// 		Content: combinedData,
// 	},
// }

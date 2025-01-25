package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

type Comment struct {
	Comment             string  `json:"comment"`
	CommentID           string  `json:"comment_id"`
	ContinuationCommand *string `json:"continuationCommand"`
}

type GetCommentResponse struct {
	Response                interface{} `json:"response"`
	CommentInfo             []Comment   `json:"comment_info"`
	NextContinuationCommand *string     `json:"nextContinuationCommand"`
}

func GetComments(url string) ([]string, error) {
	fmt.Printf("Processing url -> %s\n", url)
	if !strings.Contains(url, "www.youtube.com") {
		return nil, errors.New("invalid url: please enter a valid youtube short url")
	}
	jsonStr, err := initialRequest(url)
	if err != nil {
		return nil, errors.New("error during initial request")
	}

	ccStack := []string{}
	token := gjson.Get(*jsonStr, "engagementPanels.0.engagementPanelSectionListRenderer.header.engagementPanelTitleHeaderRenderer.menu.sortFilterSubMenuRenderer.subMenuItems.0.serviceEndpoint.continuationCommand.token")
	if !token.Exists() {
		return nil, errors.New("error while extracting the token from jsonStr")
	}
	ccStack = append(ccStack, token.String())

	comments := []string{}
	iteration := 0

	for {
		iteration++
		if iteration == 30 || len(ccStack) == 0 {
			break
		}

		token := ccStack[len(ccStack)-1]
		ccStack = ccStack[:len(ccStack)-1]
		commentResponse, err := getCommentsRequest(token)
		if err != nil {
			fmt.Printf("error while getComments\n")
		}

		for _, comment := range commentResponse.CommentInfo {
			comments = append(comments, comment.Comment)
		}
	}

	return comments, nil
}

func getCommentsRequest(token string) (*GetCommentResponse, error) {

	// Make the GET request
	url := "https://www.youtube.com/youtubei/v1/browse?prettyPrint=false"
	data := fmt.Sprintf(`{"context":{"client":{"clientName":"WEB","clientVersion":"2.20240731.04.00"}},"continuation":"%s"}`, token)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		fmt.Printf("Error while creating the POST request: %v\n", err)
		return nil, errors.New("error while creating the GET request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "www.youtube.com")
	req.Header.Set("Referer", "www.youtube.com")

	client := &http.Client{}
	// Send the request
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making the POST request: %v\n", err)
		return nil, errors.New("error making the GET request")
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading the response body: %v\n", err)
		return nil, errors.New("error reading the response body")
	}

	continuations := gjson.Get(string(body), "onResponseReceivedEndpoints.1.reloadContinuationItemsCommand.continuationItems")
	if !continuations.Exists() {
		continuations = gjson.Get(string(body), "onResponseReceivedEndpoints.0.appendContinuationItemsAction.continuationItems")
		if !continuations.Exists() {
			fmt.Printf("no continuation found\n")
		}
	}

	token2 := gjson.Get(continuations.String(), "20.continuationItemRenderer.continuationEndpoint.continuationCommand.token")
	if !token2.Exists() {
		fmt.Printf("token not found \n")
	}
	nextToken := token2.String()

	var comments []Comment
	mutations := gjson.Get(string(body), "frameworkUpdates.entityBatchUpdate.mutations")
	mutations.ForEach(func(_, mutation gjson.Result) bool {
		commentEntityPayload := mutation.Get("payload.commentEntityPayload")
		if commentEntityPayload.Exists() {
			// Extract comment properties
			commentId := commentEntityPayload.Get("properties.commentId").String()
			content := commentEntityPayload.Get("properties.content.content").String()

			var continuationCommand *string
			// Find continuation command by checking both endpoints
			appendContinuationItemsAction := gjson.Get(string(body), "onResponseReceivedEndpoints.0.appendContinuationItemsAction.continuationItems")
			if appendContinuationItemsAction.Exists() {
				appendContinuationItemsAction.ForEach(func(_, item gjson.Result) bool {
					targetId := item.Get("commentThreadRenderer.replies.commentRepliesRenderer.targetId").String()
					if strings.Contains(targetId, commentId) {
						token := item.Get("commentThreadRenderer.replies.commentRepliesRenderer.contents.0.continuationItemRenderer.continuationEndpoint.continuationCommand.token").String()
						continuationCommand = &token
						return false // stop searching
					}
					return true
				})
			}

			if continuationCommand == nil {
				// Check second endpoint for reloadContinuationItemsCommand
				reloadContinuationItemsCommand := gjson.Get(string(body), "onResponseReceivedEndpoints.1.reloadContinuationItemsCommand.continuationItems")
				if reloadContinuationItemsCommand.Exists() {
					reloadContinuationItemsCommand.ForEach(func(_, item gjson.Result) bool {
						targetId := item.Get("commentThreadRenderer.replies.commentRepliesRenderer.targetId").String()
						if strings.Contains(targetId, commentId) {
							token := item.Get("commentThreadRenderer.replies.commentRepliesRenderer.contents.0.continuationItemRenderer.continuationEndpoint.continuationCommand.token").String()
							continuationCommand = &token
							return false // stop searching
						}
						return true
					})
				}
			}

			// Add comment to the list
			comments = append(comments, Comment{
				Comment:             content,
				CommentID:           commentId,
				ContinuationCommand: continuationCommand,
			})
		}
		return true
	})

	// Send full response
	// getCommentsResponse := GetCommentResponse{
	// 	Response: string(body),
	// 	CommentInfo: comments,
	// 	NextContinuationCommand: &nextToken,
	// }

	getCommentsResponse := GetCommentResponse{
		Response:                nil,
		CommentInfo:             comments,
		NextContinuationCommand: &nextToken,
	}

	return &getCommentsResponse, errors.New("error reading the response body")
}

func initialRequest(url string) (*string, error) {

	// Make the GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error while creating the GET request: %v\n", err)
		return nil, errors.New("error while creating the GET request")
	}
	req.Header.Set("Origin", "www.youtube.com")
	req.Header.Set("Referer", "www.youtube.com")

	client := &http.Client{}
	// Send the request
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making the GET request: %v\n", err)
		return nil, errors.New("error making the GET request")
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading the response body: %v\n", err)
		return nil, errors.New("error reading the response body")
	}

	// Extract the json string within the html
	jsonStr, err := extractJsonFromHtml(string(body))
	if err != nil {
		fmt.Printf("Error reading the response body: %v\n", err)
		return nil, errors.New("error reading the response body")
	}

	return jsonStr, nil
}

func extractJsonFromHtml(body string) (*string, error) {
	cleanedResponse := strings.ReplaceAll(body, "\n", "")
	start := strings.Index(cleanedResponse, "var ytInitialData = ")
	if start == -1 {
		fmt.Printf("Error 'ytInitialData' not found \n")
		return nil, errors.New("error 'ytInitialData not found'")
	}
	// Adjust start index to skip "var ytInitialData = "
	start += len("var ytInitialData = ")
	end := strings.Index(cleanedResponse[start:], ";</script>")
	if end == -1 {
		fmt.Printf("Error : Closing tag not found\n")
		return nil, errors.New("error 'closing tag not found'")
	}
	// Extract the JSON string and parse it
	jsonStr := cleanedResponse[start : start+end]
	return &jsonStr, nil
}

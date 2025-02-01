package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

const (
	totalIterations    = 3
	filterCommentsFlag = false
)

func ExtractData(url string) (*ExtractDataResponse, error) {
	fmt.Printf("Processing url -> %s\n", url)
	if !strings.Contains(url, "www.youtube.com") {
		return nil, errors.New("invalid url: please enter a valid youtube short url")
	}
	initialReponse, err := initialRequest(url)
	if err != nil {
		return nil, errors.New("error during initial request")
	}

	subtitles, err := getSubtitlesData(initialReponse.InitialPlayerResponse)
	if err != nil {
		fmt.Printf("error while getting subtitles data: err %v\n", err)
		return nil, errors.New("error while getting subtitles data")
	}

	comments, err := getCommentsData(initialReponse.InitialData)
	if err != nil {
		fmt.Printf("error while getting comments data: err %v\n", err)
		return nil, errors.New("error while getting comments data")
	}

	extractDataResponse := ExtractDataResponse{
		Comments: comments,
		Subtitles: subtitles,
	}

	return &extractDataResponse, nil
}

func getCommentsData(data *string) ([]string, error) {

	ccStack := []string{}
	token := gjson.Get(*data, "engagementPanels.0.engagementPanelSectionListRenderer.header.engagementPanelTitleHeaderRenderer.menu.sortFilterSubMenuRenderer.subMenuItems.0.serviceEndpoint.continuationCommand.token")
	if !token.Exists() {
		return nil, errors.New("error while extracting the token from jsonStr")
	}
	ccStack = append(ccStack, token.String())

	comments := []string{}
	iteration := 0

	for {
		iteration++
		if iteration == totalIterations || len(ccStack) == 0 {
			break
		}

		token := ccStack[len(ccStack)-1]
		ccStack = ccStack[:len(ccStack)-1]
		commentResponse, err := getCommentsRequest(token)
		if err != nil {
			fmt.Printf("error while getComments\n")
		}

		if commentResponse.NextContinuationCommand != nil && *commentResponse.NextContinuationCommand != "" {
			ccStack = append(ccStack, *commentResponse.NextContinuationCommand)
		}

		if filterCommentsFlag {
			filterComments(commentResponse.CommentInfo, &comments, &ccStack)
		} else {
			for _, comment := range commentResponse.CommentInfo {
				comments = append(comments, comment.Comment)
			}
		}
	}

	return comments, nil

}

func getSubtitlesData(data *string) (*string, error) {

	baseUrl := gjson.Get(*data, "captions.playerCaptionsTracklistRenderer.captionTracks.0.baseUrl")
	if !baseUrl.Exists() {
		return nil, errors.New("error while extracting the token from jsonStr")
	}
	timedtextUrl := baseUrl.String()

	return &timedtextUrl, nil

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

	return &getCommentsResponse, nil
}

func initialRequest(url string) (*InitialReponse, error) {

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
	initialData, err := extractJsonFromHtml(string(body), "var ytInitialData = ", ";</script>")
	if err != nil {
		fmt.Printf("Error reading the response body: %v\n", err)
		return nil, errors.New("error reading the response body")
	}

	initialPlayerResponse, err := extractJsonFromHtml(string(body), "var ytInitialPlayerResponse = ", ";var")
	if err != nil {
		fmt.Printf("Error reading the response body: %v\n", err)
		return nil, errors.New("error reading the response body")
	}

	initialReponse := InitialReponse{
		InitialData:           initialData,
		InitialPlayerResponse: initialPlayerResponse,
	}

	return &initialReponse, nil
}

func filterComments(inputComments []Comment, comments *[]string, ccStack *[]string) {
	movieMentionRegex := regexp.MustCompile(`(?:\b(movie|film|cinema|show|series|watched|saw|seen|about|called|name)\s+|(?:"([^"]+)"|'([^']+)'))`)
	askingForMovieRegex := regexp.MustCompile(`\b(what.s|which|can|anybody|please)\s+(movie|show|series|scene|film|is|was|this|that|it|tell)\??|name\s+(of|this|that)\s+(movie|show|series)\??`)

	for _, comment := range inputComments {
		if movieMentionRegex.MatchString(comment.Comment) {
			*comments = append(*comments, comment.Comment)
		}
	}

	for _, comment := range inputComments {
		if askingForMovieRegex.MatchString(comment.Comment) {
			if comment.ContinuationCommand != nil && *comment.ContinuationCommand != "" {
				*ccStack = append(*ccStack, *comment.ContinuationCommand)
			}
		}
	}
}

func extractJsonFromHtml(body string, startStr string, endStr string) (*string, error) {
	cleanedResponse := strings.ReplaceAll(body, "\n", "")
	// start := strings.Index(cleanedResponse, "var ytInitialData = ")
	start := strings.Index(cleanedResponse, startStr)
	if start == -1 {
		fmt.Printf("Error '%s' not found \n", startStr)
		return nil, errors.New(fmt.Sprintf("error start '%s' not found \n", startStr))
	}
	// Adjust start index to skip "var ytInitialData = "
	start += len(startStr)
	end := strings.Index(cleanedResponse[start:], endStr)
	if end == -1 {
		fmt.Printf("Error '%s' not found \n", endStr)
		return nil, errors.New(fmt.Sprintf("error endStr '%s' not found \n", endStr))
	}
	// Extract the JSON string and parse it
	jsonStr := cleanedResponse[start : start+end]
	return &jsonStr, nil
}

// https://www.youtube.com/api/timedtext?v=16tWbpk8sws\u0026ei=UdmdZ9tswOyDxQ_kja7pCg\u0026caps=asr\u0026opi=112496729\u0026xoaf=5\u0026xosf=1\u0026hl=en\u0026ip=0.0.0.0\u0026ipbits=0\u0026expire=1738423233\u0026sparams=ip,ipbits,expire,v,ei,caps,opi,xoaf\u0026signature=67FBC37F84BD59F60AF32D37E6BA529C0C3EAD8F.D1FA5026DC0D84C5006FD9F72DC7C80519E9A490\u0026key=yt8\u0026kind=asr\u0026lang=en

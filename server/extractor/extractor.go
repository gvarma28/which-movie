package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gvarma28/which-movie/server/utils"
	"github.com/tidwall/gjson"
)

const (
	totalIterations    = 3
	filterCommentsFlag = false
)

func ExtractData(url string) (*ExtractDataResponse, error) {
	fmt.Printf("Processing url -> %s\n", url)
	if !strings.Contains(url, "youtube.com") {
		return nil, errors.New("invalid url: please enter a valid youtube short url")
	}
	initialReponse, err := initialRequest(url)
	if err != nil {
		fmt.Printf("error while getting initialResponse: err %v\n", err)
		return nil, errors.New("error during initial request")
	}

	subtitles, err := getSubtitlesData(initialReponse.InitialPlayerResponse)
	if err != nil {
		fmt.Printf("error while getting subtitles data: err %v\n", err)
	}

	comments, err := getCommentsData(initialReponse.InitialData)
	if err != nil {
		fmt.Printf("error while getting comments data: err %v\n", err)
		return nil, errors.New("error while getting comments data")
	}

	title, err := getTitleData(initialReponse.InitialPlayerResponse)
	if err != nil {
		fmt.Printf("error while getting title data: err %v\n", err)
		return nil, errors.New("error while getting title data")
	}

	extractDataResponse := ExtractDataResponse{
		Comments:  comments,
		Subtitles: subtitles,
		Title:     title,
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

func getSubtitlesData(data map[string]any) (*string, error) {
	timedtextUrl := utils.FindInJSON(data, "captions", "playerCaptionsTracklistRenderer", "captionTracks", "0", "baseUrl").(string)
	body, err := getSubtitlesRequest(timedtextUrl)
	if err != nil {
		fmt.Printf("error at getSubtitlesRequest %v \n", err)
		return nil, fmt.Errorf("error at getSubtitlesRequest %v", err)
	}

	var subtitles []string
	eventJSON, err := utils.ConvertToJSON(body)
	if err != nil {
		return nil, err
	}
	eventArr := eventJSON["events"].([]any)
	for _, eventObj := range eventArr {
		subtitleArr, ok := eventObj.(map[string]any)["segs"].([]any)
		if !ok {
			continue
		}
		for _, subtitleObj := range subtitleArr {
			subtitle, ok := subtitleObj.(map[string]any)["utf8"].(string)
			if !ok {
				continue
			}
			subtitles = append(subtitles, subtitle)
		}
	}
	subtitle := strings.Join(subtitles, "")

	return &subtitle, nil
}

func getTitleData(data map[string]any) (*string, error) {
	baseData := data["videoDetails"].(map[string]any)
	title := baseData["title"].(string)
	return &title, nil
}

func getSubtitlesRequest(baseUrl string) ([]byte, error) {

	parsedURL, err := url.Parse(baseUrl)
	if err != nil {
		fmt.Printf("error parsing the url %v", err)
	}

	queryParams := parsedURL.Query()
	queryParams.Set("fmt", "json3")
	queryParams.Set("lang", "en")

	parsedURL.RawQuery = queryParams.Encode()

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		fmt.Printf("Error while creating the GET request: %v\n", err)
		return nil, errors.New("error while creating the GET request")
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

	// save html to file
	// err = os.WriteFile("output.txt", body, 0644)
	// if err != nil {
	// 	panic(err)
	// }

	endStrArr := []string{";</script>"}
	// Extract the json string within the html
	initialData, err := extractJsonFromHtml(string(body), "var ytInitialData = ", endStrArr)
	if err != nil {
		fmt.Printf("Error reading the response body: %v\n", err)
		return nil, errors.New("error reading the response body")
	}

	endStrArr = append(endStrArr, ";var")
	initialPlayerResponse, err := extractJsonFromHtml(string(body), "var ytInitialPlayerResponse = ", endStrArr)
	if err != nil {
		fmt.Printf("Error reading the response body: %v\n", err)
		return nil, errors.New("error reading the response body")
	}
	jsonInitialPlayerResponse, err := utils.ConvertToJSON([]byte(*initialPlayerResponse))
	if err != nil {
		return nil, errors.New("error while converting json to map")
	}

	initialReponse := InitialReponse{
		InitialData:           initialData,
		InitialPlayerResponse: jsonInitialPlayerResponse,
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

func extractJsonFromHtml(body string, startStr string, endStr []string) (*string, error) {
	cleanedResponse := strings.ReplaceAll(body, "\n", "")
	// start := strings.Index(cleanedResponse, "var ytInitialData = ")
	start := strings.Index(cleanedResponse, startStr)
	if start == -1 {
		fmt.Printf("Error '%s' not found \n", startStr)
		return nil, fmt.Errorf("error start '%s' not found", startStr)
	}
	// Adjust start index to skip "var ytInitialData = "
	start += utf8.RuneCountInString(startStr)

	minLastIdx := math.MaxInt
	for _, v := range endStr {
		end := strings.Index(cleanedResponse[start:], v)
		if end == -1 {
			continue
		}
		if minLastIdx > end {
			minLastIdx = end
		}
	}

	// Extract the JSON string and parse it
	jsonStr := cleanedResponse[start : start+minLastIdx]
	return &jsonStr, nil
}

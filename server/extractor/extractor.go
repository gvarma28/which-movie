package extractor

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gvarma28/which-movie/server/utils"
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

func getCommentsData(data map[string]any) ([]string, error) {
	ccStack := []string{}
	token := utils.FindInJSON(data, "engagementPanels", "0", "engagementPanelSectionListRenderer", "header", "engagementPanelTitleHeaderRenderer", "menu", "sortFilterSubMenuRenderer", "subMenuItems", "0", "serviceEndpoint", "continuationCommand", "token")
	if token == nil {
		return nil, errors.New("error while extracting the token from jsonStr")
	}
	ccStack = append(ccStack, token.(string))

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

	body, err := utils.GetRequest(parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func getCommentsRequest(token string) (*GetCommentResponse, error) {
	// Make the GET request
	url := "https://www.youtube.com/youtubei/v1/browse?prettyPrint=false"
	data := fmt.Sprintf(`{"context":{"client":{"clientName":"WEB","clientVersion":"2.20240731.04.00"}},"continuation":"%s"}`, token)

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Origin"] = "www.youtube.com"
	headers["Referer"] = "www.youtube.com"

	body, err := utils.PostRequest(url, headers, []byte(data))
	if err != nil {
		return nil, err
	}

	jsonBody, err := utils.ConvertToJSON(body)
	if err != nil {
		fmt.Printf("Error while converting json to map: %v\n", err)
		return nil, err
	}

	token2 := utils.FindInJSON(jsonBody, "onResponseReceivedEndpoints", "1", "reloadContinuationItemsCommand", "continuationItems", "20", "continuationItemRenderer", "continuationEndpoint", "continuationCommand", "token")
	if token2 == nil {
		token2 = utils.FindInJSON(jsonBody, "onResponseReceivedEndpoints", "0", "appendContinuationItemsAction", "continuationItems", "20", "continuationItemRenderer", "continuationEndpoint", "continuationCommand", "token")
	}
	nextToken := token2.(string)

	var comments []Comment
	mutationsArr := utils.FindInJSON(jsonBody, "frameworkUpdates", "entityBatchUpdate", "mutations").([]any)
	for _, mutation := range mutationsArr {
		commentEntityPayload := utils.FindInJSON(mutation, "payload", "commentEntityPayload")
		if commentEntityPayload == nil {
			continue
		}
		commentId := utils.FindInJSON(commentEntityPayload, "properties", "commentId").(string)
		content := utils.FindInJSON(commentEntityPayload, "properties", "content", "content").(string)

		var continuationCommand *string

		reloadContinuationItemsCommand := utils.FindInJSON(jsonBody, "onResponseReceivedEndpoints", "1", "reloadContinuationItemsCommand", "continuationItems")
		if reloadContinuationItemsCommand == nil {
			continue
		}
		reloadContinuationItemsCommandArr := reloadContinuationItemsCommand.([]any)
		for _, item := range reloadContinuationItemsCommandArr {
			continuationCommandRes, err := extractContinuationCommand(item, commentId)
			if err != nil {
				continue
			}
			continuationCommand = continuationCommandRes
			break
		}

		if continuationCommand == nil {
			appendContinuationItemsAction := utils.FindInJSON(jsonBody, "onResponseReceivedEndpoints", "0", "appendContinuationItemsAction", "continuationItems")
			if appendContinuationItemsAction != nil {
				appendContinuationItemsActionArr := appendContinuationItemsAction.([]any)
				for _, item := range appendContinuationItemsActionArr {
					continuationCommandRes, err := extractContinuationCommand(item, commentId)
					if err != nil {
						continue
					}
					continuationCommand = continuationCommandRes
					break
				}
			}
		}
	
		cleanComment := cleanComment(content)

		// Add comment to the list
		comments = append(comments, Comment{
			Comment:             cleanComment,
			CommentID:           commentId,
			ContinuationCommand: continuationCommand,
		})

	}

	getCommentsResponse := GetCommentResponse{
		Response:                nil,
		CommentInfo:             comments,
		NextContinuationCommand: &nextToken,
	}

	return &getCommentsResponse, nil
}

func initialRequest(url string) (*InitialReponse, error) {
	headers := make(map[string]string)
	headers["Origin"] = "www.youtube.com"
	headers["Referer"] = "www.youtube.com"
	body, err := utils.GetRequest(url, nil)
	if err != nil {
		return nil, err
	}

	endStrArr := []string{";</script>"}
	// Extract the json string within the html
	initialData, err := extractJsonFromHtml(string(body), "var ytInitialData = ", endStrArr)
	if err != nil {
		fmt.Printf("Error reading the response body: %v\n", err)
		return nil, err
	}
	jsonInitialData, err := utils.ConvertToJSON([]byte(*initialData))
	if err != nil {
		fmt.Printf("Error while converting json to map: %v\n", err)
		return nil, err
	}

	endStrArr = append(endStrArr, ";var")
	initialPlayerResponse, err := extractJsonFromHtml(string(body), "var ytInitialPlayerResponse = ", endStrArr)
	if err != nil {
		fmt.Printf("Error reading the response body: %v\n", err)
		return nil, err
	}
	jsonInitialPlayerResponse, err := utils.ConvertToJSON([]byte(*initialPlayerResponse))
	if err != nil {
		fmt.Printf("Error while converting json to map: %v\n", err)
		return nil, err
	}

	initialReponse := InitialReponse{
		InitialData:           jsonInitialData,
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

func extractContinuationCommand(item any, commentId string) (*string, error) {
	commentRepliesRenderer := utils.FindInJSON(item, "commentThreadRenderer", "replies", "commentRepliesRenderer")
	if commentRepliesRenderer == nil {
		return nil, errors.New("commentRepliesRenderer doesn't exist in object")
	}
	commentRepliesRendererMap := commentRepliesRenderer.(map[string]any)
	targetId := commentRepliesRendererMap["targetId"].(string)
	if strings.Contains(targetId, commentId) {
		token := utils.FindInJSON(commentRepliesRenderer, "contents", "0", "continuationItemRenderer", "continuationEndpoint", "continuationCommand", "token")
		if token == nil {
			return nil, errors.New("token doesn't exist in object")
		}
		tokenStr := token.(string)
		return &tokenStr, nil
	}
	return nil, errors.New("couldn't find a match")
}

func cleanComment(comment string) string {
    // Convert to lowercase
    comment = strings.ToLower(comment)

    // Remove emojis and special characters
    // Regex to remove emojis and other non-alphanumeric characters except spaces
    reg := regexp.MustCompile(`[^\p{L}\p{N}\s]`)
    comment = reg.ReplaceAllString(comment, "")

    // Remove extra whitespaces
    comment = regexp.MustCompile(`\s+`).ReplaceAllString(comment, " ")

    // Remove common filler words and noise
    removeWords := map[string]bool{
        "i": true, "the": true, "a": true, "an": true, "and": true, 
        "or": true, "but": true, "in": true, "on": true, "at": true, 
        "to": true, "for": true, "of": true, "with": true, "by": true, 
        "from": true, "up": true, "about": true, "lol": true, "wow": true, 
        "omg": true, "crazy": true, "like": true, "just": true, "so": true, 
        "his": true, "her": true, "their": true, "its": true, "edit": true,
    }

    // Split into words
    words := strings.Fields(comment)
    
    // Filter out remove words and trim
    var cleanedWords []string
    for _, word := range words {
        // Skip if word is in remove list or too short
        if !removeWords[word] && len(word) > 1 {
            cleanedWords = append(cleanedWords, word)
        }
    }

    // Limit to first 20 words
    if len(cleanedWords) > 20 {
        cleanedWords = cleanedWords[:20]
    }

    return strings.TrimSpace(strings.Join(cleanedWords, " "))
}
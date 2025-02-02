package extractor

type Comment struct {
	Comment             string  `json:"comment"`
	CommentID           string  `json:"comment_id"`
	ContinuationCommand *string `json:"continuationCommand"`
}

type GetCommentResponse struct {
	Response                interface{} `json:"response"`
	CommentInfo             []Comment   `json:"comment_info"`
	NextContinuationCommand *string     `json:"next_continuation_command"`
}

type InitialReponse struct {
	InitialData           *string `json:"initial_data"`
	InitialPlayerResponse *string `json:"initial_player_response"`
}

type ExtractDataResponse struct {
	Comments  []string `json:"comments"`
	Subtitles *string  `json:"subtitles"`
	Title     *string  `json:"title"`
}

package processor

// Open-AI Request Body
type RequestBody struct {
	Model    string     `json:"model"`
	Messages []Messages `json:"messages"`
}

type Messages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Valid Open-AI Response
type APIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Failed Open-AI Response
type APIErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param"` // Use pointer for nullable values
	Code    *string `json:"code"`  // Use pointer for nullable values
}

// open-ai response
type MagicResult struct {
	Results []MovieResult `json:"results"`
}

type MovieResult struct {
	MovieName        string `json:"movie_name"`
	Year             int    `json:"year"`
	ShortDescription string `json:"short_description"`
}

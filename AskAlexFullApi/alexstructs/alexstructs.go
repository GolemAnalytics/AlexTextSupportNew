package alexstructs

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`

}

type PayLoad struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Model string `json:"model"`
	Messages []Message `json:"messages"`
	Temperature int `json:"temperature"`
	Max_Tokens int `json:"max_tokens"`
	Top_p int `json:"top_p"`
	Frequency_Penalty int `json:"frequency_penalty"`
	Presence_Penalty int `json:"presence_penalty"`
}

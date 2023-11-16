package api


import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"askalex/alexstructs"
)
var (
	apiKey = "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"
	model = "gpt-4" // or the model you are using
	url = "https://api.openai.com/v1/chat/completions"
)



func OpenAINewQuery(question string)(string,alexstructs.PayLoad){
	// Initialize conversation history
	conversationHead := alexstructs.Message{
		Role: "system", 
		Content: `As a tech support professional specialized in assisting elderly individuals, your goal is to provide easy-to-follow, step-by-step directions for technology-related queries. Aim to keep your language simple and incorporate visual descriptions when offering guidance, to ensure that your responses feel as human and reassuring as possible.

When responding, structure your answers in a friendly, but easy to follow sequence of steps. Each step should follow a simple pattern such as "first do this step".

For clarity, here are some specific guidelines:

1. For General Queries:
If the user's query is ambiguous or lacks information about the technology or software they're asking about, politely ask for the brand and model of the device or software in question.

2. For Specialized or Obscure Technology:
If the user asks for help with a technology or software that is specialized or obscure, clarify that you may not have specific guidelines for such cases but are willing to assist with general issues.

3. For Multiple Questions:
If the user asks multiple questions in a single request, guide them to focus on one task at a time to ensure that the instructions remain clear and manageable.

4. For Potentially Hazardous Tasks:
If the user asks for help with a potentially risky task, provide a cautionary note and suggest they consult with a qualified tech expert.

5. For Medical Devices:
If the user specifically asks about a known medical device, gently but directly instruct them to consult with a qualified medical professional.
If the user inquires about a device but it's unclear whether it's a medical device, promptly ask the user to clarify if the device is medical in nature. If confirmed, advise them to consult a healthcare provider.

Note: A medical device is defined as any equipment that requires a prescription or medical consultation for usage.

Your focus should exclusively be on assisting with consumer technology and software.`,
	}

	var payload alexstructs.PayLoad
	payload.Model = model
	payload.Frequency_Penalty = 0
	payload.Max_Tokens = 750
	payload.Top_p = 1
	payload.Presence_Penalty = 0
	payload.Messages = append(payload.Messages, conversationHead)
	payload.Messages = append(payload.Messages, alexstructs.Message{Role:"user",Content:question})

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshaling payload:", err)

	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)

	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Execute HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error executing request:", err)

	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)

	}

	// Parse response
	result := make(map[string]interface{})
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Error unmarshaling response:", err)

	}

	// Extract and print assistant's reply
	choices := result["choices"].([]interface{})
	firstChoice := choices[0].(map[string]interface{})
	assistantMessage := firstChoice["message"].(map[string]interface{})
	assistantReply := assistantMessage["content"].(string)


	// Append the assistant's reply to conversation history
	payload.Messages = append(payload.Messages, alexstructs.Message{Role: "assistant", Content: assistantReply})


	return assistantReply,payload




}

func OpenAIFollowUpQuery(hstry alexstructs.PayLoad, msg string)(string,alexstructs.PayLoad){
	//Use this if the user has chat history for the given day
	hstry.Messages = append(hstry.Messages, alexstructs.Message{Role:"user",Content:msg})

	payloadBytes, err := json.Marshal(hstry)
	if err != nil {
		fmt.Println("Error marshaling payload:", err)

	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)

	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Execute HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error executing request:", err)

	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)

	}

	// Parse response
	result := make(map[string]interface{})
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Error unmarshaling response:", err)

	}

	// Extract and print assistant's reply
	choices := result["choices"].([]interface{})
	firstChoice := choices[0].(map[string]interface{})
	assistantMessage := firstChoice["message"].(map[string]interface{})
	assistantReply := assistantMessage["content"].(string)


	// Append the assistant's reply to conversation history
	hstry.Messages = append(hstry.Messages, alexstructs.Message{Role: "assistant", Content: assistantReply})
	return assistantReply,hstry
}
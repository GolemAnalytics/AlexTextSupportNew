//TODO: Make end points to do the following
// - Validate if user is a valid user
// - validate if user text is the first of the day or if there is text history
// - DB management
// - validate if we have seen this question before
// - make a call to open ai api
// - send message to customer with answer
// - DB management
// - add new customers upon stripe payment
// - DB management
// - remove users upon request
// - DB management

//Project structure
// dir: API
// api_post.go
// dir: DB
// db_new.go
// db_update.go
// db.go
// dir: UTILS
// alex_utils.go
// file: askalex.go
// mod and tidy

package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/joho/godotenv"

	"encoding/json"

	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"database/sql"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/webhook"
	"github.com/twilio/twilio-go"
	twilioapi "github.com/twilio/twilio-go/rest/api/v2010"
	"github.com/twilio/twilio-go/twiml"

	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"

	"bytes"
)

var (
	Db *sqlx.DB
	err error
	apiKey = os.Getenv("OPENAIKEY")
	model = "gpt-4" // or the model you are using
	url = "https://api.openai.com/v1/chat/completions"
	alertmsg = `I hope you're doing well. This is Alex from AskAlex. We're reaching out because your loved one might be going through a difficult time and could really use your support. A quick call to check in on them could mean a lot. Your care and attention are invaluable. Thanks for being there for them.`

)

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

func init()  {
	godotenv.Load(".env")
}

func main() {
	router := gin.Default()
	router.POST("/incomingmsg",IncomingMsgHandler)
	router.POST("/askalexaddnewuser",NewUserHandler)
	router.POST("/askalexrenewuser",RenewUserHandler)
	router.POST("/askalexuserevents",UserAccountHandler)
	err := router.Run(":8080")
	if err != nil{
		fmt.Println(err)
	}
}


func AskAlexQuestionCheck()(bool,string){
	//a function that checks if the input question has an answer
	//if this is a new question, add question to db, return false and an empty string
	//if this is not a new question, return true and the answer
	var PastQuestion string
	var QuestionPresent bool


	return QuestionPresent,PastQuestion
}

func SendMsgHandler(msg, number string){
	// a function to send messages using twilio number and library 
		// Find your Account SID and Auth Token at twilio.com/console
	// and set the environment variables. See http://twil.io/secure
	
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACCOUNT_SID"),
		Password: os.Getenv("TWILIO_AUTH_TOKEN"),
	})
	params := &twilioapi.CreateMessageParams{}
	params.SetBody(msg)
	params.SetFrom(os.Getenv("TWILIO_NUMBER"))
	params.SetTo(number)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		if resp.Sid != nil {
			fmt.Println(*resp.Sid)
		} else {
			fmt.Println(resp.Sid)
		}
	}
}


func IncomingMsgHandler(c *gin.Context){
	// a function to take in a new message from twilio object
	var msg = &twiml.MessagingMessage{}

	body := c.PostForm("Body")
	number := c.PostForm("From")

	switch AskAlexStatusCheck(number){
	case true:
		if AskAlexFollowUpQuestion(number){
			//if true, then this is a follow up question
			hstry := AskAlexGetQuestions(number)
			asstiantResp,jsonObj := OpenAIFollowUpQuery(hstry,body,number)
			SendMsgHandler(asstiantResp,number)
			//marshal results into json for storage
			AskAlexUpdateQuestion(number,jsonObj)

		}else{
			//not a follow up quetsion
			asstiantResp,jsonObj := OpenAINewQuery(body,number)
			//send the results from the first response
			SendMsgHandler(asstiantResp,number)
			//marshal results into json for storage
			AskAlexSaveQuestion(number,jsonObj)


		}
	case false:
		SendMsgHandler("Hey, it looks like you arent signed up for our service. Please visit www.golemanalytics.com/askalex to get the help you need",number)
	}
	// Generate and return our TwiML response
	twiml, _ := twiml.Messages([]twiml.Element{msg})
	c.Header("Content-Type", "text/xml")
	c.String(http.StatusOK, twiml)
} 

func NewUserHandler(c *gin.Context){
	//a function that handles a successful stripe payment for ask alex
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Error reading request body"})
		return
	}

	// Verify the event by checking its signature
	event, err := webhook.ConstructEvent(payload, c.Request.Header.Get("Stripe-Signature"),  os.Getenv("EndPointSecret"))
	
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error verifying webhook signature"})
		return
	}

	// Handle the event
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing checkout.session.async_payment_succeede event"})
			return
		}

		NumberToAdd := session.CustomFields[0].Numeric.Value
		parentNumber := session.Customer.Phone
		ParentAccountId := session.Customer.ID
		AskAlexNewMember(ParentAccountId,"+1"+NumberToAdd,parentNumber)
		SendMsgHandler("Hello and welcome! I'm Alex, your friendly tech support guide at Golem Analytics. If you're setting up any devices or need help signing up for a service like Netflix, please know that I'm here just for you. Don't worry if technology seems a bit tricky – I'll be with you at every step, offering easy-to-follow, patient guidance. Should you have any questions or face any challenges, feel free to reach out to me. Together, we'll make sure everything works smoothly for you. Your comfort and confidence in using our services is my utmost priority!","+1"+NumberToAdd)

	default:
		c.JSON(http.StatusOK, gin.H{"message": "Unhandled event type"})
	}
}

func RenewUserHandler(c *gin.Context){
	//a function that handles a successful stripe payment for ask alex
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Error reading request body"})
		return
	}

	// Verify the event by checking its signature
	event, err := webhook.ConstructEvent(payload, c.Request.Header.Get("Stripe-Signature"), os.Getenv("EndPointSecret"))
	
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error verifying webhook signature"})
		return
	}

	// Handle the event
	switch event.Type {
	case "invoice.payment_succeeded":
		var session stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing checkout.session.async_payment_succeede event"})
			return
		}

		NumberToAdd := session.Charge.Invoice.CustomFields[0].Value		
		AskAlexReNewMember("+1"+NumberToAdd)

	default:
		c.JSON(http.StatusOK, gin.H{"message": "Unhandled event type"})
	}
}

func UserAccountHandler(c *gin.Context){
		//a function that handles a successful stripe payment for ask alex
		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Error reading request body"})
			return
		}
	
		// Verify the event by checking its signature
		event, err := webhook.ConstructEvent(payload, c.Request.Header.Get("Stripe-Signature"), os.Getenv("EndPointSecret"))
		
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error verifying webhook signature"})
			return
		}
	
		// Handle the event
		switch event.Type {
		case "customer.subscription.deleted","customer.subscription.updated":
			var session stripe.Subscription
			err := json.Unmarshal(event.Data.Raw, &session)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing checkout.session.async_payment_succeede event"})
				return
			}
	
			ParentID := session.Customer.ID
			AskAlexCancelMember(ParentID)

		case "invoice.payment_succeeded":
			var session stripe.Invoice
			err := json.Unmarshal(event.Data.Raw, &session)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing checkout.session.async_payment_succeede event"})
				return
			}
			ParentID := session.Customer.ID
			AskAlexCancelMember(ParentID)
		default:
			c.JSON(http.StatusOK, gin.H{"message": "Unhandled event type"})
		}
}

func Connect() {
	connstring := os.Getenv("DBConnection")

	Db, err = sqlx.Connect("pgx",connstring)
	if err != nil{
		fmt.Println(err)
	}

}


func AskAlexStatusCheck(number string)bool{
	//a function to check user status
	//returns false if user is not in the database or has status of false. Else true
	var UserStatus bool
	Connect()
	defer Db.Close()

	queryString := fmt.Sprintf(`SELECT "Status" FROM public."AlexStatus" WHERE "Number" = '%s'`,number)
	err := Db.QueryRow(queryString).Scan(&UserStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			// No rows were returned
			UserStatus = false
		} else {
			// An error occurred during the query
			UserStatus = false
		}
	} 
	return UserStatus
}

func AskAlexGetParentNumber(customerNumber string)string{

	var ParentNumber string
	Connect()
	defer Db.Close()

	queryString := fmt.Sprintf(`SELECT "ParentNumber" FROM public."AlexStatus" WHERE "Number" = '%s'`,customerNumber)
	err := Db.QueryRow(queryString).Scan(&ParentNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			// No rows were returned
			ParentNumber = ""
		} else {
			// An error occurred during the query
			ParentNumber = ""
		}
	} 
	return ParentNumber
}

func AskAlexFollowUpQuestion(number string)bool{
	//check if the user has a query for the day
	TodaysDate := time.Now().Format("2006-01-02")
	var QueryDate string

	var status bool
	Connect()
	defer Db.Close()
	query := fmt.Sprintf(`SELECT TO_CHAR(MAX("Date"), 'YYYY-MM-DD') FROM public."AlexHstry" WHERE "Number" = '%s'`,number)
	err := Db.QueryRow(query).Scan(&QueryDate)
	if err != nil{
		status = false
	}else{
		status = QueryDate == TodaysDate
	}

	return status
}

func AskAlexSaveQuestion(number string, obj PayLoad){
	currentDate := time.Now().Format("2006-01-02") 
	// Serialize the PayLoad to JSON
	jsonData, err := json.Marshal(obj)
	if err != nil {
		fmt.Println(err)
	}
	Connect()
	defer Db.Close()
	// Insert the JSON data into the database
	_, err = Db.Exec(`INSERT INTO public."AlexHstry" ("Date","Number","Hstry") VALUES ($1,$2,$3)`, currentDate,number,jsonData)
	if err != nil {
		fmt.Println(err)
	}

}

func AskAlexUpdateQuestion(number string, obj PayLoad){
	currentDate := time.Now().Format("2006-01-02") 
	// Serialize the PayLoad to JSON
	jsonData, err := json.Marshal(obj)
	if err != nil {
		fmt.Println(err)
	}
	Connect()
	defer Db.Close()
	// Insert the JSON data into the database


	_, err = Db.Exec(`UPDATE public."AlexHstry" SET "Hstry"=$1 WHERE "Number"=$2 AND "Date" = $3`,jsonData,number,currentDate)
	if err != nil {
		fmt.Println(err)
	}

}
		
func AskAlexGetQuestions(number string)PayLoad{
	var datareturn string
	var masterPayLoad PayLoad
	Connect()
	defer Db.Close()
	currentDate := time.Now().Format("2006-01-02") 

	query := fmt.Sprintf(`SELECT "Hstry" FROM public."AlexHstry" WHERE "Number" ='%s' AND "Date" = '%s'`,number,currentDate)

	err := Db.QueryRow(query).Scan(&datareturn)
	if err != nil{
		fmt.Println(err)
	}

	jsonErr := json.Unmarshal([]byte(datareturn),&masterPayLoad)
	if jsonErr != nil{
		fmt.Println(jsonErr)
	}
	return masterPayLoad
}
		
func AskAlexNewMember(number,ID,parentNumber string){
	Connect()
	defer Db.Close()
	currentDate := time.Now().Format("2006-01-02") 
	endMonthDate := time.Now().AddDate(0,1,0).Format("2006-01-02") 
	_, err = Db.Exec(`INSERT INTO public."AlexStatus" ("Number", "Status", "JoinDate", "EndDate","ParentID","ParentNumber") VALUES ($1,$2,$3,$4,$5,$6)`,number,true,currentDate,endMonthDate,ID,parentNumber)
	if err != nil{
		fmt.Println(err)
	}
}
		
func AskAlexReNewMember(number string){
	Connect()
	defer Db.Close()

	endMonthDate := time.Now().AddDate(0,1,0).Format("2006-01-02") 
	updatewuery := fmt.Sprintf(`UPDATE public."AlexStatus" SET "EndDate"=%s WHERE "Number"=%s;`,endMonthDate,number)
	_, err = Db.Exec(updatewuery)
	if err != nil{
		fmt.Println(err)
	}
}

func AskAlexCancelMember(ParentID string){
	Connect()
	defer Db.Close()

	endMonthDate := time.Now().Format("2006-01-02") 
	_, err = Db.Exec(`UPDATE public."AlexStatus" SET "EndDate"=$1 "Status"=$2 WHERE "ParentID"=$3;`,endMonthDate,false,ParentID)
	if err != nil{
		fmt.Println(err)
	}
}

func OpenAINewQuery(question,incoming_number string)(string,PayLoad){
	// Initialize conversation history
	conversationHead := Message{
		Role: "system", 
		Content: `As a tech support professional dedicated to assisting elderly individuals, my primary focus is on helping with consumer technology and software queries. Please note, I do not provide driving or navigation directions. My aim is to offer simple, step-by-step instructions that are easy to follow. Here's how I can assist you:General Queries: If your question is unclear or lacks specific details about the technology, kindly provide the brand and model of the device or software.Specialized Technology: For less common technology or software, I can offer general advice, as I may not have detailed guidelines for all types of technology.Handling Multiple Questions: If you have several questions, let's tackle them one at a time. This approach ensures clear and manageable guidance.Potentially Risky Tasks: Should you inquire about a task that seems hazardous, I'll caution you and recommend consulting a tech expert.Medical Devices: For queries about medical devices (equipment requiring a prescription or medical consultation), please consult a healthcare professional. If it's unclear whether your device is medical, I'll ask for clarification and advise accordingly. Responses begin with ASSISTANT_START and end with ASSISTANT_END. For medical-related queries, indicate true or false in MEDICAL_QUERY_START and MEDICAL_QUERY_END. For risky tasks, indicate true or false in ALERT_START' and ALERT_END. My responses should always follow the format of ASSISTANT_START the answer to the query ASSISTANT_END MEDICAL_QUERY_START true if a medical device question is present else false MEDICAL_QUERY_END ALERT_START true if potentially risky task asked else false ALERT_END. Responses are limited to 1600 characters. If I lack knowledge, I will provide instructions on how to Google a solution.`,
	}

	var payload PayLoad
	payload.Model = model
	payload.Frequency_Penalty = 0
	payload.Max_Tokens = 1600
	payload.Top_p = 1
	payload.Presence_Penalty = 0
	payload.Messages = append(payload.Messages, conversationHead)
	payload.Messages = append(payload.Messages, Message{Role:"user",Content:question})

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
	assistantReply,medical_query,alert := OpenAIAssistantResponseParse(assistantMessage["content"].(string))
	assistaneReplyRaw := assistantMessage["content"].(string)

	if strings.ToLower(medical_query)=="true" || strings.ToLower(alert)=="true"{
		parentNumber := AskAlexGetParentNumber(incoming_number)
		SendMsgHandler(alertmsg,parentNumber)	
		}

	// Append the assistant's reply to conversation history
	payload.Messages = append(payload.Messages, Message{Role: "assistant", Content: assistaneReplyRaw})

	return assistantReply,payload




}
		
func OpenAIFollowUpQuery(hstry PayLoad, msg, incoming_number string)(string,PayLoad){
	//Use this if the user has chat history for the given day
	hstry.Messages = append(hstry.Messages, Message{Role:"user",Content:msg})

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
	assistantReply,medical_query,alert := OpenAIAssistantResponseParse(assistantMessage["content"].(string))
	assistaneReplyRaw := assistantMessage["content"].(string)

	if strings.ToLower(medical_query)=="true" || strings.ToLower(alert)=="true"{
		parentNumber := AskAlexGetParentNumber(incoming_number)
		SendMsgHandler(alertmsg,parentNumber)	
		}


	// Append the assistant's reply to conversation history
	hstry.Messages = append(hstry.Messages, Message{Role: "assistant", Content: assistaneReplyRaw})
	return assistantReply,hstry
}

func OpenAIAssistantResponseParse(response string)(string,string,string){
	//a function that takes in an OpenAI and vertix response and parse out the parts that are needed
	var (
		assistant string
		medical_query string
		alert string
		assistant_regex = regexp.MustCompile(`(ASSISTANT_START(?s).+ASSISTANT_END)`)
		medical_regex = regexp.MustCompile(`(MEDICAL_QUERY_START(?s).+MEDICAL_QUERY_END)`)
		alert_regex = regexp.MustCompile(`(ALERT_START(?s).+ALERT_END)`)
	)

	assistant =strings.Replace(strings.Replace(assistant_regex.FindString(response),"ASSISTANT_START","",-1),"ASSISTANT_END","",-1)
	medical_query = strings.Replace(strings.Replace(medical_regex.FindString(response),"MEDICAL_QUERY_START","",-1),"MEDICAL_QUERY_END","",-1)
	alert = strings.Replace(strings.Replace(alert_regex.FindString(response),"ALERT_START","",-1),"ALERT_END","",-1)


	return assistant,medical_query,alert
}

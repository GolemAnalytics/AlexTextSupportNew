package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"askalex/api"
	"askalex/db"

	"github.com/gin-gonic/gin"

	"github.com/twilio/twilio-go"
	twilioapi "github.com/twilio/twilio-go/rest/api/v2010"
	"github.com/twilio/twilio-go/twiml"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/webhook"
)





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
	
	
	fmt.Println(os.Getenv("TWILIO_ACCOUNT_SID"))
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
	fmt.Print(c)
	body := c.PostForm("Body")
	number := c.PostForm("From")

	switch db.AskAlexStatusCheck(number){
	case true:
		if db.AskAlexFollowUpQuestion(number){
			//if true, then this is a follow up question
			hstry := db.AskAlexGetQuestions(number)
			asstiantResp,jsonObj := api.OpenAIFollowUpQuery(hstry,body)
			SendMsgHandler(asstiantResp,number)
			//marshal results into json for storage
			db.AskAlexSaveQuestion(number,jsonObj)

		}else{
			//not a follow up quetsion
			asstiantResp,jsonObj := api.OpenAINewQuery(body)
			//send the results from the first response
			SendMsgHandler(asstiantResp,number)
			//marshal results into json for storage
			db.AskAlexSaveQuestion(number,jsonObj)


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
        db.AskAlexNewMember("+1"+NumberToAdd)

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
        db.AskAlexReNewMember("+1"+NumberToAdd)

    default:
        c.JSON(http.StatusOK, gin.H{"message": "Unhandled event type"})
    }
}
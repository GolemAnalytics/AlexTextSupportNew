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
	"askalex/utils"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
)

func init()  {
	godotenv.Load(".env")
}

func main() {
	router := gin.Default()
	router.POST("/incomingmsg",utils.IncomingMsgHandler)
	router.POST("/askalexaddnewuser",utils.NewUserHandler)
	router.POST("/askalexrenewuser",utils.RenewUserHandler)
	err := router.Run("localhost:8080")
	if err != nil{
		fmt.Println(err)
	}
}


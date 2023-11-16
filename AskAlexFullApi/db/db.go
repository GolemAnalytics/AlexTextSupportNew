package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"askalex/alexstructs"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

var (
	Db *sqlx.DB
	err error

)



func Connect() {
	connstring := os.Getenv("DBConnection")
	fmt.Println(connstring)
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

func AskAlexSaveQuestion(number string, obj alexstructs.PayLoad){
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

func AskAlexGetQuestions(number string)alexstructs.PayLoad{
	var datareturn alexstructs.PayLoad
	Connect()
	defer Db.Close()
	currentDate := time.Now().Format("2006-01-02") 

	query := fmt.Sprintf(`SELECT "Hstry" FROM public."AlexHstry" WHERE "Number" ='%s' AND "Date" = '%s'`,number,currentDate)

	err := Db.QueryRow(query).Scan(&datareturn)
	if err != nil{
		fmt.Println(err)
	}
	return datareturn
}

func AskAlexNewMember(number string){
	Connect()
	defer Db.Close()
	currentDate := time.Now().Format("2006-01-02") 
	endMonthDate := time.Now().AddDate(0,1,0).Format("2006-01-02") 
	_, err = Db.Exec(`INSERT INTO public."AlexStatus" ("Number", "Status", "JoinDate", "EndDate") VALUES ($1,$2,$3,$4)`,number,true,currentDate,endMonthDate)
	if err != nil{
		fmt.Println(err)
	}
}

func AskAlexReNewMember(number string){
	Connect()
	defer Db.Close()

	endMonthDate := time.Now().AddDate(0,1,0).Format("2006-01-02") 
	updatewuery := fmt.Sprintf(`UPDATE public."AlexStatus" SET "EndDate"=%s WHERE "Number"==%s;`,endMonthDate,number)
	_, err = Db.Exec(updatewuery)
	if err != nil{
		fmt.Println(err)
	}
}
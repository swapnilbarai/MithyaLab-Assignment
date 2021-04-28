package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)
// schema used for sql query for signup  and format at which data recieved in json format but TotalPoint is not needed
type User struct {
	Email       string `json:"Email"`
	Password    string `json:"Password"`
	Totalpoint  int    `json:"Totalpoint"`
	ReferenceId string `json:"ReferenceId"`
}

var db *sql.DB  // pointer to mysqlserver connection

func SignUp(c *gin.Context) {
	var Newuser User
	if err := c.ShouldBindJSON(&Newuser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}  
	fmt.Println(Newuser)

	//condition for checking emaail or password not entered by signup user
	if Newuser.Email == "" || Newuser.Password == "" {
		c.JSON(http.StatusOK, gin.H{
			"Message": "Please Enter the Email or Password",
		})
		return
	}

	// sql query to check if email already exit
	results, err := db.Query("SELECT Password FROM signup where Email = ?", Newuser.Email)
	defer results.Close()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cnt := 0
	for results.Next() {
		cnt++

	}
	if cnt != 0 {
		c.JSON(http.StatusOK, gin.H{
			"Message": "Please Enter the another Email because following email aready used",
		})
		return
	}

	var Email string
	var ReferenceId string

	// this is done to check the coupon enter by user exit or not 
	CheckReferenceId, err := db.Query("SELECT Email FROM signup where ReferenceId = ?", Newuser.ReferenceId)
	for CheckReferenceId.Next() {

		err := CheckReferenceId.Scan(&Email)
		ReferenceId = Newuser.ReferenceId
		fmt.Println(Email)
		if err != nil {
			panic(err.Error())
		}
	}
	if len(Email) == 0{
		c.JSON(http.StatusOK, gin.H{
			"Message": "Please Enter the valid Coupon for SignUp!",
		})
		return
	}
	//creating RefererId for New User
	Newuser.ReferenceId = "234" + Newuser.Email + "#321"
	Newuser.Totalpoint = 5 // score of Sigup user
	insert, err := db.Query("INSERT INTO signup VALUES ( ?, ?,?,?)", Newuser.Email, Newuser.Password, Newuser.Totalpoint, Newuser.ReferenceId)
	if err != nil {
		panic(err.Error())
	}
	defer insert.Close()
	//updating the score of Reference User
	insert1, err := db.Query("UPDATE signup SET Totalpoint=Totalpoint+10 where Email= ?", Email)
	defer insert1.Close()
	if err != nil {
		panic(err.Error())
	}
	// inserting referId used by signup user and its email which will be used for querying no. user use specific ReferId
	insert3, err := db.Query("INSERT INTO Refer VALUES (?,?)", ReferenceId,Newuser.Email)
	defer insert3.Close()
	if err != nil {
		panic(err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"Message": "You are Succusfully signUp", "ReferenceId": Newuser.ReferenceId})
	return
}

func UserWithRefer(c *gin.Context) {
	ReferId := c.Param("ReferId") // refer id from url

	results, err := db.Query("SELECT Email FROM Refer where ReferenceId = ?", ReferId) 
	//checking referid used by which user in mysql
	defer results.Close()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var Email string
	var Emails []string
	for results.Next() {
		results.Scan(&Email)
		Emails = append(Emails, Email)
	}
  
	// checking condition if referid does not exit or not used by any user
	if len(Emails) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": " No User has used the following RefererId or RefererId does not exit!",
		})
		return
	}
	// returning users
	c.JSON(http.StatusOK, gin.H{"Users": Emails})
	return
}

func PointsWithReferer(c *gin.Context) {
	Email := c.Param("UserId") //getting userid from url 
	results, err := db.Query("SELECT Totalpoint FROM signup where Email = ?", Email)
	//querying in mysql for score 
	defer results.Close()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var Score int64
	IsExitRefID := 0
	for results.Next() {
		results.Scan(&Score)
		IsExitRefID++
	}
	// condition if email doesnot exit or reference id is not used while making account 
	if IsExitRefID == 0 || Score==0 {
		c.JSON(http.StatusOK, gin.H{"Score": 0})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Score": Score - 5})
	return
}

func main() {

	var err error
	db, err = sql.Open("mysql", "root:Swapnil@123@tcp(127.0.0.1:3306)/website")

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Sucessfully connected to mysql server")
	}

	defer db.Close()
	r := gin.Default()
	v := r.Group("/")
	{
		v.POST("/Signup", SignUp)
		v.GET("/UserWithReferCode/:ReferId/", UserWithRefer)
		v.GET("/PointsbyReffering/:UserId/", PointsWithReferer)
	}
	r.Run()

}

//improvement are needed like adding logger , creating component , using environment variables but as as the time is less
//it is sufficient 

// create two tables first one signup+-------------+--------------+------+-----+---------+-------+
//| Field       | Type         | Null | Key | Default | Extra |
//+-------------+--------------+------+-----+---------+-------+
//| Email       | varchar(255) | NO   | PRI | NULL    |       |
//| Password    | varchar(500) | NO   |     | NULL    |       |
//| Totalpoint  | int          | NO   |     | NULL    |       |
//| ReferenceId | varchar(255) | NO   |     | NULL    |       |
//+-------------+--------------+------+-----+---------+-------+
// second one refer
//

//+-------------+--------------+------+-----+---------+-------+
//| Field       | Type         | Null | Key | Default | Extra |
//+-------------+--------------+------+-----+---------+-------+
//| ReferenceId | varchar(255) | NO   | PRI | NULL    |       |
//| User        | varchar(255) | NO   |     | NULL    |       |
//+-------------+--------------+------+-----+---------+-------+

package models

import (
	"fmt"

	// "github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
)

type User struct {
	Id       int    `json:"UserId"`
	Name     string `json:"UserName"`
	Email    string `json:"UserEmail"`
	Password string `json:"UserPassword"`
}

// Get users (test)
func FindALLUsers(c *fiber.Ctx)error {
	db, _ := ConnectToMYSQL()
	allUsers, err := db.Query("SELECT * FROM member;")
	if err != nil {
		fmt.Printf("查詢資料庫失敗，原因為：%v\n", err)
	}
	defer allUsers.Close()
	fmt.Println("測試全部的使用者")
	fmt.Println(allUsers)
	// 宣告一個User結構的slice，這個 slice 將會存放查詢得到的所有使用者
	// users := []User{}
	var users []User
	for allUsers.Next() {
		// 宣告一個 User 類型的變數，這個變數將會存放每一筆查詢得到的使用者資料。
		// user := User{}
		var user User
		err := allUsers.Scan(&user.Id, &user.Name, &user.Email, &user.Password)
		if err != nil {
			fmt.Printf("讀取資料失敗，原因為：%v\n", err)
		}
		users = append(users, user)
	}

	return c.JSON(users)


	// [
	// {
	// "UserId": 1,
	// "UserName": "dylan",
	// "UserEmail": "dylan@gmail.com",
	// "UserPassword": "1qaz"
	// },
	// {
	// "UserId": 2,
	// "UserName": "curry",
	// "UserEmail": "curry@warrior.com",
	// "UserPassword": "2wsx"
	// }
	// ]

}

// Post user
// func PostUser(c *gin.Context) {
// 	user := User{}
// 	err := c.BindJSON(&user)
// 	if err != nil {
// 		c.JSON(http.StatusNotAcceptable, "Error : "+err.Error())
// 	}
// 	c.JSON(http.StatusOK, "Successfully posted")
// 	// 資料庫動作
// }

//Delete user  登出會員
// func LogOutUser(c *gin.Context){
// }

//Put user  登入會員
// func PutUser(c *gin.Context){
// }

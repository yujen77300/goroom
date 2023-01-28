package models

import (
	"fmt"
	"log"
	"time"

	// "github.com/gin-gonic/gin"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

type User struct {
	Id       int    `json:"id"`
	Name     string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Get all users (test)
func FindALLUsers(c *fiber.Ctx) error {
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

// New user
func NewUser(c *fiber.Ctx) error {
	signUpInfo := User{}
	if err := c.BodyParser(&signUpInfo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true, "message": "Cannot parse data to struct",
		})
	}
	fmt.Println("測試一下新增的資料")
	fmt.Println(signUpInfo.Id)
	fmt.Println(signUpInfo.Name)
	fmt.Println(signUpInfo.Password)
	fmt.Println(signUpInfo.Email)

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln(err)
	}
	db, _ := ConnectToMYSQL()
	row, _ := db.Query("SELECT email FROM member WHERE email = ?;", signUpInfo.Email)
	fmt.Println(row)
	var signUpMember []User
	for row.Next() {
		var member User
		if dberr := row.Scan(&member.Email); dberr != nil {
			fmt.Printf("scan failed, err:%v\n", dberr)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "scan failed"})
		}
		signUpMember = append(signUpMember, member)
	}
	row.Close()
	fmt.Println("測試一下")
	fmt.Println(signUpMember)
	if len(signUpMember) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": true, "message": "Email is already registered"})
	} else {
		result, err := db.Exec("INSERT INTO member(username,email,password) values(?,?,?);", signUpInfo.Name, signUpInfo.Email, signUpInfo.Password)
		if err != nil {
			fmt.Printf("建立檔案失敗，原因是：%v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Internal Server Error"})
		}
		rowsaffected, err := result.RowsAffected()
		if err != nil {
			fmt.Printf("Get RowsAffected failed,err:%v", err)
		}
		fmt.Println("Affected rows:", rowsaffected)
	}

	return c.JSON(fiber.Map{
		"ok": true,
	})

}

// Get user 取得當前登入資料
func GetUser(c *fiber.Ctx) error {

	livedToken := c.Cookies("MyJWT")
	if len(livedToken) == 0 {
		return c.JSON(fiber.Map{
			"data": "no data",
		})
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatalln(err)
		}

		JWTSECRECT := viper.GetString("JWTSECRECT")
		secretKey := JWTSECRECT
		secret := []byte(secretKey)
		token, err := jwt.Parse(livedToken, func(token *jwt.Token) (interface{}, error) {
			// check token signing method
			return secret, nil
		})

		if err != nil {
			return nil
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			fmt.Println("解析jwt")
			fmt.Println(claims)
			fmt.Printf("資料型態 : %T\n", claims)
			email := claims["email"].(string)
			id := claims["id"].(float64)
			name := claims["name"].(string)
			// use interface{}to store any type
			memberData := map[string]interface{}{
				"id":    id,
				"name":  name,
				"email": email,
			}
			return c.JSON(fiber.Map{
				"data": memberData,
			})
		} else {
			log.Printf("Invalid JWT Token")
			return nil
		}
	}
}

// Delete user登出會員
func SignOutUser(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:    "MyJWT",
		Expires: time.Now().Add(-(time.Hour * 1)),
	})
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"ok": true})
}

// Put user  登入會員
func PutUser(c *fiber.Ctx) error {
	signInInfo := User{}
	if err := c.BodyParser(&signInInfo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true, "message": "Cannot parse data to struct",
		})
	}

	db, _ := ConnectToMYSQL()
	row, _ := db.Query("SELECT id,username,email FROM member WHERE email = ? AND password=?;", signInInfo.Email, signInInfo.Password)
	fmt.Println(row)
	defer row.Close()
	// 建立一個slice來儲存資料
	var members []User
	for row.Next() {
		var member User
		if dberr := row.Scan(&member.Id, &member.Name, &member.Email); dberr != nil {
			fmt.Printf("scan failed, err:%v\n", dberr)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "scan failed"})
		}
		members = append(members, member)
	}

	fmt.Println("測試一下搜尋結果")
	fmt.Println(members)

	if len(members) == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Email or password is incorrect",
		})
	} else {
		token := jwt.New(jwt.SigningMethodHS256)
		// 存在token裡面的body
		claims := token.Claims.(jwt.MapClaims)
		claims["id"] = members[0].Id
		claims["name"] = members[0].Name
		claims["email"] = members[0].Email
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatalln(err)
		}

		JWTSECRECT := viper.GetString("JWTSECRECT")
		jwtToken, err := token.SignedString([]byte(JWTSECRECT))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "jwt token error",
			})
		}
		fmt.Println(jwtToken)
		c.Cookie(&fiber.Cookie{
			Name:     "MyJWT",
			Value:    jwtToken,
			HTTPOnly: true,
			Expires:  time.Now().Add(time.Hour * 24 * 7),
		})
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"ok": true})
	}

}

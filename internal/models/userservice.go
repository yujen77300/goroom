package models

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

type signInUser struct {
	Id        int    `json:"id"`
	Name      string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	AvatarUrl string `json:"avatarurl"`
}

type UserAvatar struct {
	Email     string `json:"email"`
	AvatarUrl string `json:"avatarurl"`
}

type PcpAvatar struct {
	UserId    int    `json:"id"`
	Email     string `json:"email"`
	AvatarUrl string `json:"avatarurl"`
}

var UserEmail string


func NewUser(c *fiber.Ctx) error {
	signUpInfo := User{}
	if err := c.BodyParser(&signUpInfo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true, "message": "Cannot parse data to struct",
		})
	}

	userData := fiber.Map{
		"newUserEmail":    signUpInfo.Email,
		"newUserPassword": signUpInfo.Password,
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln(err)
	}
	if emailValidation(signUpInfo.Email) && pwdValidation(signUpInfo.Password) {
		db, _ := ConnectToMYSQL()
		row, _ := db.Query("SELECT email FROM member WHERE email = ?;", signUpInfo.Email)
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

		if len(signUpMember) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": true, "message": "Email is already registered"})
		} else {
			result, err := db.Exec("INSERT INTO member(username,email,password) values(?,?,?);", signUpInfo.Name, signUpInfo.Email, signUpInfo.Password)
			defer db.Close()
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
			"ok":           true,
			"newUsesrInfo": userData,
		})
	} else if emailValidation(signUpInfo.Email) && !pwdValidation(signUpInfo.Password) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid password, at least 8 characters, one number and one English letter are required"})
	} else if !emailValidation(signUpInfo.Email) && pwdValidation(signUpInfo.Password) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid email"})
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid email and password"})
	}

}

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
			return secret, nil
		})

		if err != nil {
			return nil
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			email := claims["email"].(string)
			id := claims["id"].(float64)
			name := claims["name"].(string)
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

func SignOutUser(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:    "MyJWT",
		Expires: time.Now().Add(-(time.Hour * 1)),
	})
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"ok": true})
}

func PutUser(c *fiber.Ctx) error {
	signInInfo := signInUser{}
	if err := c.BodyParser(&signInInfo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true, "message": "Cannot parse data to struct",
		})
	}

	db, _ := ConnectToMYSQL()
	row, _ := db.Query("SELECT id,username,email,avatar_url FROM member WHERE email = ? AND password=?;", signInInfo.Email, signInInfo.Password)
	defer row.Close()

	var members []signInUser
	for row.Next() {
		var member signInUser
		if dberr := row.Scan(&member.Id, &member.Name, &member.Email,&member.AvatarUrl); dberr != nil {
			fmt.Printf("scan failed, err:%v\n", dberr)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "scan failed"})
		}
		members = append(members, member)
	}

	if len(members) == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Email or password is incorrect",
		})
	} else {
		token := jwt.New(jwt.SigningMethodHS256)
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
		c.Cookie(&fiber.Cookie{
			Name:     "MyJWT",
			Value:    jwtToken,
			HTTPOnly: true,
			Expires:  time.Now().Add(time.Hour * 24 * 7),
		})

		redisConn := RedisDefaultPool.Get()
		defer redisConn.Close()
		redisConn.Do("HMSET", members[0].Id, "cacheName",members[0].Name, "cacheAvatarUrl",members[0].AvatarUrl)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"ok": true})
	}

}

func GetAvatar(c *fiber.Ctx) error {
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
			return secret, nil
		})

		if err != nil {
			return nil
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			email := claims["email"].(string)
			memberData := map[string]interface{}{
				"email": email,
			}
			value := memberData["email"]
			UserEmail = value.(string)
		} else {
			log.Printf("Invalid JWT Token")
		}
	}
	db, _ := ConnectToMYSQL()
	row, err := db.Query("SELECT email,avatar_url FROM member WHERE email = ?;", UserEmail)
	if err != nil {
		fmt.Printf("Database query failed, error：%v\n", err)
	}
	defer row.Close()
	defer db.Close()
	var userAvatar []UserAvatar
	for row.Next() {
		var user UserAvatar
		if dberr := row.Scan(&user.Email, &user.AvatarUrl); dberr != nil {
			fmt.Printf("scan failed, err:%v\n", dberr)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "scan failed"})
		}
		userAvatar = append(userAvatar, user)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"user": UserEmail, "userAvatar": userAvatar[0].AvatarUrl})
}

func GetPcpAvatar(c *fiber.Ctx) error {
	pcpEmail := c.Params("userEmail")
	pcpEmail = strings.TrimLeft(pcpEmail, ":")
	db, _ := ConnectToMYSQL()
	row, err := db.Query("SELECT id,email,avatar_url FROM member WHERE email = ?;", pcpEmail)
	if err != nil {
		fmt.Printf("Database query failed, error：%v\n", err)
	}
	defer row.Close()
	defer db.Close()
	var pcpAvatar []PcpAvatar
	for row.Next() {
		var pcp PcpAvatar
		if dberr := row.Scan(&pcp.UserId, &pcp.Email, &pcp.AvatarUrl); dberr != nil {
			fmt.Printf("scan failed, err:%v\n", dberr)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "scan failed"})
		}
		pcpAvatar = append(pcpAvatar, pcp)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"pcpUserId": pcpAvatar[0].UserId, "pcpEmail": pcpEmail, "pcpAvatarUrl": pcpAvatar[0].AvatarUrl})
}

func UpdateAvatar(c *fiber.Ctx) error {
	region, bucketName, client := ConnectToAWS()
	userNow := c.FormValue("accountEmail")
	file, err := c.FormFile("avatarFile")
	if err != nil {
		fmt.Println(err)
	}
	contentDisposition := file.Header["Content-Disposition"][0]
	fileFormat := strings.Split(contentDisposition, ".")[1]
	fileFormat = strings.Replace(fileFormat, "\"", "", -1)

	var alphabet []rune = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	alphabetSize := len(alphabet)
	var sb strings.Builder

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 20; i++ {
		ch := alphabet[rand.Intn(alphabetSize)]
		sb.WriteRune(ch)
	}
	randomFileName := sb.String()
	fileName := randomFileName + "." + fileFormat

	newFile, err := file.Open()
	if err != nil {
		fmt.Println(err)
	}

	_, error := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   newFile,
		ACL:    "public-read",
	})
	if error != nil {
		fmt.Printf("Couldn't upload file, Here's why: %v\n", error)
	}

	url := "https://" + bucketName + ".s3." + region + ".amazonaws.com/" + fileName
	fmt.Println(url)
	cloudFrontUrl := "https://d1uumvm880lnxp.cloudfront.net/" + fileName

	db, _ := ConnectToMYSQL()
	_, updateErr := db.Exec("UPDATE member SET avatar_url = ? WHERE email= ?;", cloudFrontUrl, userNow)
	if updateErr != nil {
		return updateErr
	}
	defer db.Close()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"userEmail": userNow, "newAvatarUrl": cloudFrontUrl})

}

func emailValidation(email string) bool {
	emailRegex := regexp.MustCompile(`^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,4})+$`)
	result := emailRegex.MatchString(email)
	return result
}

func pwdValidation(pwd string) bool {
	hasNum := false
	hasLetter := false

	for _, r := range pwd {
		if unicode.IsDigit(r) {
			hasNum = true
		} else if unicode.IsLetter(r) {
			hasLetter = true
		}
	}

	if hasNum && hasLetter && utf8.RuneCountInString(pwd) >= 8 {
		return true
	} else {
		return false
	}
}

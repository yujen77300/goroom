package models

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"database/sql"
)

func ConnectToMYSQL() (*sql.DB, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}

	const (
		NETWORK = "tcp"
		PORT    = 3306
	)

	USERNAME := viper.GetString("USERNAME")
	PASSWORD := viper.GetString("PASSWORD")
	DATABASE := viper.GetString("DATABASE")
	SERVER := viper.GetString("SERVER")


	conn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", USERNAME, PASSWORD, NETWORK, SERVER, PORT, DATABASE)
	db, err := sql.Open("mysql", conn)
	if err != nil {
		return nil, fmt.Errorf("開啟 MySQL 連線發生錯誤，原因為： %v", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("資料庫連線錯誤，原因為： %v", err)
	}
	return db, nil
}

